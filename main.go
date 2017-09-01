package main

import (
	"fmt"
	"os"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/microkit/transaction"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"
	"github.com/spf13/viper"

	"github.com/giantswarm/draughtsman-operator/flag"
	"github.com/giantswarm/draughtsman-operator/server"
	"github.com/giantswarm/draughtsman-operator/service"
	"github.com/giantswarm/draughtsman-operator/service/configurer/configmap"
	"github.com/giantswarm/draughtsman-operator/service/configurer/secret"
	"github.com/giantswarm/draughtsman-operator/service/eventer/github"
	"github.com/giantswarm/draughtsman-operator/service/installer/helm"
	"github.com/giantswarm/draughtsman-operator/service/notifier/slack"
)

var (
	description string     = "The draughtsman-operator is an in-cluster agent that handles Helm based deployments on behalf of the draughtsmantpr."
	f           *flag.Flag = flag.New()
	gitCommit   string     = "n/a"
	name        string     = "draughtsman-operator"
	source      string     = "https://github.com/giantswarm/draughtsman-operator"
)

func main() {
	err := mainWithError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", microerror.Mask(err)))
	}
}

func mainWithError() error {
	var err error

	// Create a new logger which is used by all packages.
	var newLogger micrologger.Logger
	{
		loggerConfig := micrologger.DefaultConfig()
		loggerConfig.IOWriter = os.Stdout
		newLogger, err = micrologger.New(loggerConfig)
		if err != nil {
			panic(err)
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	newServerFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			serviceConfig := service.DefaultConfig()

			serviceConfig.Flag = f
			serviceConfig.Logger = newLogger
			serviceConfig.Viper = v

			serviceConfig.Description = description
			serviceConfig.GitCommit = gitCommit
			serviceConfig.Name = name
			serviceConfig.Source = source

			newService, err = service.New(serviceConfig)
			if err != nil {
				panic(err)
			}
			go newService.Boot()
		}

		var storage microstorage.Storage
		{
			storage, err = memory.New(memory.DefaultConfig())
			if err != nil {
				panic(err)
			}
		}

		var transactionResponder transaction.Responder
		{
			c := transaction.DefaultResponderConfig()
			c.Logger = newLogger
			c.Storage = storage

			transactionResponder, err = transaction.NewResponder(c)
			if err != nil {
				panic(err)
			}
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			serverConfig := server.DefaultConfig()

			serverConfig.MicroServerConfig.Logger = newLogger
			serverConfig.MicroServerConfig.ServiceName = name
			serverConfig.MicroServerConfig.TransactionResponder = transactionResponder
			serverConfig.MicroServerConfig.Viper = v
			serverConfig.Service = newService

			newServer, err = server.New(serverConfig)
			if err != nil {
				panic(err)
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		commandConfig := command.DefaultConfig()

		commandConfig.Logger = newLogger
		commandConfig.ServerFactory = newServerFactory

		commandConfig.Description = description
		commandConfig.GitCommit = gitCommit
		commandConfig.Name = name
		commandConfig.Source = source

		newCommand, err = command.New(commandConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Configurer.ConfigMap.Key, "values", "Key in configmap holding values data.")
	daemonCommand.PersistentFlags().String(f.Service.Configurer.ConfigMap.Name, "draughtsman-values-configmap", "Name of configmap holding values data.")
	daemonCommand.PersistentFlags().String(f.Service.Configurer.ConfigMap.Namespace, "draughtsman", "Namespace of configmap holding values data.")
	daemonCommand.PersistentFlags().String(f.Service.Configurer.File.Path, "", "Path to values file.")
	daemonCommand.PersistentFlags().String(f.Service.Configurer.Secret.Key, "values", "Key in secret holding values data.")
	daemonCommand.PersistentFlags().String(f.Service.Configurer.Secret.Name, "draughtsman-values-secret", "Name of secret holding values data.")
	daemonCommand.PersistentFlags().String(f.Service.Configurer.Secret.Namespace, "draughtsman", "Namespace of secret holding values data.")
	daemonCommand.PersistentFlags().String(f.Service.Configurer.Types, string(configmap.ConfigurerType)+","+string(secret.ConfigurerType), "Comma separated list of configurers to use for configuration management.")

	daemonCommand.PersistentFlags().String(f.Service.Eventer.Environment, "", "Environment name that draughtsman is running in.")
	daemonCommand.PersistentFlags().String(f.Service.Eventer.GitHub.OAuthToken, "", "OAuth token for authenticating against GitHub. Needs 'repo_deployment' scope.")
	daemonCommand.PersistentFlags().String(f.Service.Eventer.GitHub.Organisation, "", "Organisation under which to check for deployments.")
	daemonCommand.PersistentFlags().String(f.Service.Eventer.Type, string(github.GithubEventerType), "Which eventer to use for event management.")

	daemonCommand.PersistentFlags().Duration(f.Service.HTTPClient.Timeout, 10*time.Second, "Timeout for HTTP requests.")

	daemonCommand.PersistentFlags().String(f.Service.Installer.Helm.HelmBinaryPath, "/bin/helm", "Path to Helm binary. Needs CNR registry plugin installed.")
	daemonCommand.PersistentFlags().String(f.Service.Installer.Helm.Organisation, "", "Organisation of Helm CNR registry.")
	daemonCommand.PersistentFlags().String(f.Service.Installer.Helm.Password, "", "Password for Helm CNR registry.")
	daemonCommand.PersistentFlags().String(f.Service.Installer.Helm.Registry, "quay.io", "URL for Helm CNR registry.")
	daemonCommand.PersistentFlags().String(f.Service.Installer.Helm.Username, "", "Username for Helm CNR registry.")
	daemonCommand.PersistentFlags().String(f.Service.Installer.Type, string(helm.HelmInstallerType), "Which installer to use for installation management.")

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	daemonCommand.PersistentFlags().String(f.Service.Notifier.Slack.Channel, "", "Channel to post Slack notifications to.")
	daemonCommand.PersistentFlags().String(f.Service.Notifier.Slack.Emoji, ":older_man:", "Emoji to use for Slack notifications.")
	daemonCommand.PersistentFlags().String(f.Service.Notifier.Slack.Token, "", "Token to post Slack notifications with.")
	daemonCommand.PersistentFlags().String(f.Service.Notifier.Slack.Username, "draughtsman", "Username to post Slack notifications with.")
	daemonCommand.PersistentFlags().String(f.Service.Notifier.Type, string(slack.SlackNotifierType), "Which notifier to use for notification management.")

	newCommand.CobraCommand().Execute()

	return nil
}
