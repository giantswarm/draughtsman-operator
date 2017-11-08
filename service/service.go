// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/giantswarm/draughtsmantpr"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/logresource"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/operatorkit/tpr"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/draughtsman-operator/flag"
	"github.com/giantswarm/draughtsman-operator/service/eventer"
	eventerspec "github.com/giantswarm/draughtsman-operator/service/eventer/spec"
	"github.com/giantswarm/draughtsman-operator/service/healthz"
	"github.com/giantswarm/draughtsman-operator/service/installer"
	installerspec "github.com/giantswarm/draughtsman-operator/service/installer/spec"
	"github.com/giantswarm/draughtsman-operator/service/notifier"
	notifierspec "github.com/giantswarm/draughtsman-operator/service/notifier/spec"
	"github.com/giantswarm/draughtsman-operator/service/operator"
	"github.com/giantswarm/draughtsman-operator/service/resource/project"
)

type Config struct {
	Logger micrologger.Logger

	Description string
	Flag        *flag.Flag
	GitCommit   string
	Name        string
	Source      string
	Viper       *viper.Viper
}

func DefaultConfig() Config {
	return Config{
		Logger: nil,

		Description: "",
		Flag:        nil,
		GitCommit:   "",
		Name:        "",
		Source:      "",
		Viper:       nil,
	}
}

type Service struct {
	Healthz  *healthz.Service
	Operator *operator.Operator
	Version  *version.Service

	bootOnce sync.Once
}

func New(config Config) (*Service, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var k8sClient kubernetes.Interface
	{
		k8sConfig := k8sclient.DefaultConfig()

		k8sConfig.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		k8sConfig.Logger = config.Logger
		k8sConfig.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		k8sConfig.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		k8sConfig.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		k8sConfig.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		k8sClient, err = k8sclient.New(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var httpClient *http.Client
	{
		timeout := config.Viper.GetDuration(config.Flag.Service.HTTPClient.Timeout)
		if timeout.Seconds() == 0 {
			return nil, microerror.Maskf(invalidConfigError, "http client timeout must be greater than zero")
		}

		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	var osFileSystem afero.Fs
	{
		osFileSystem = afero.NewOsFs()
	}

	var notifierService notifierspec.Notifier
	{
		notifierConfig := notifier.DefaultConfig()

		notifierConfig.Logger = config.Logger

		notifierConfig.Flag = config.Flag
		notifierConfig.Viper = config.Viper

		notifierConfig.Type = notifierspec.NotifierType(config.Viper.GetString(config.Flag.Service.Notifier.Type))

		notifierService, err = notifier.New(notifierConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var installerService installerspec.Installer
	{
		installerConfig := installer.DefaultConfig()

		installerConfig.FileSystem = osFileSystem
		installerConfig.K8sClient = k8sClient
		installerConfig.Logger = config.Logger

		installerConfig.Flag = config.Flag
		installerConfig.Type = installerspec.InstallerType(config.Viper.GetString(config.Flag.Service.Installer.Type))
		installerConfig.Viper = config.Viper

		installerService, err = installer.New(installerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var eventerService eventerspec.Eventer
	{
		eventerConfig := eventer.DefaultConfig()

		eventerConfig.HTTPClient = httpClient
		eventerConfig.Logger = config.Logger

		eventerConfig.Flag = config.Flag
		eventerConfig.Viper = config.Viper

		eventerService, err = eventer.New(eventerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var projectResource *project.Resource
	{
		projectConfig := project.DefaultConfig()

		projectConfig.Eventer = eventerService
		projectConfig.Installer = installerService
		projectConfig.Logger = config.Logger
		projectConfig.Notifier = notifierService

		projectResource, err = project.New(projectConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// We create the list of resources and wrap each resource around some common
	// resources like metrics and retry resources.
	//
	// NOTE that the retry resources wrap the underlying resources first. The
	// wrapped resources are then wrapped around the metrics resource. That way
	// the metrics also consider execution times and execution attempts including
	// retries.
	var resources []framework.Resource
	{
		resources = []framework.Resource{
			projectResource,
		}

		logWrapConfig := logresource.DefaultWrapConfig()
		logWrapConfig.Logger = config.Logger
		resources, err = logresource.Wrap(resources, logWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		retryWrapConfig := retryresource.DefaultWrapConfig()
		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.NewExponentialBackOff() }
		retryWrapConfig.Logger = config.Logger
		resources, err = retryresource.Wrap(resources, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.DefaultWrapConfig()
		metricsWrapConfig.Name = config.Name
		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	var frameworkBackOff *backoff.ExponentialBackOff
	{
		frameworkBackOff = backoff.NewExponentialBackOff()
		frameworkBackOff.MaxElapsedTime = 5 * time.Minute
	}

	var operatorFramework *framework.Framework
	{
		frameworkConfig := framework.DefaultConfig()

		frameworkConfig.BackOff = frameworkBackOff
		frameworkConfig.InitCtxFunc = initCtxFunc
		frameworkConfig.Logger = config.Logger
		frameworkConfig.Resources = resources

		operatorFramework, err = framework.New(frameworkConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newTPR *tpr.TPR
	{
		c := tpr.DefaultConfig()

		c.K8sClient = k8sClient
		c.Logger = config.Logger

		c.Description = draughtsmantpr.Description
		c.Name = draughtsmantpr.Name
		c.Version = draughtsmantpr.VersionV1

		newTPR, err = tpr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newWatcherFactory informer.WatcherFactory
	{
		zeroObjectFactory := &informer.ZeroObjectFactoryFuncs{
			NewObjectFunc:     func() runtime.Object { return &draughtsmantpr.CustomObject{} },
			NewObjectListFunc: func() runtime.Object { return &draughtsmantpr.List{} },
		}
		newWatcherFactory = informer.NewWatcherFactory(k8sClient.Discovery().RESTClient(), newTPR.WatchEndpoint(""), zeroObjectFactory)
	}

	var newInformer *informer.Informer
	{
		informerConfig := informer.DefaultConfig()

		informerConfig.BackOff = backoff.NewExponentialBackOff()
		informerConfig.WatcherFactory = newWatcherFactory

		informerConfig.ResyncPeriod = 10 * time.Second

		newInformer, err = informer.New(informerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var healthzService *healthz.Service
	{
		healthzConfig := healthz.DefaultConfig()

		healthzConfig.K8sClient = k8sClient
		healthzConfig.Logger = config.Logger

		healthzService, err = healthz.New(healthzConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorBackOff *backoff.ExponentialBackOff
	{
		operatorBackOff = backoff.NewExponentialBackOff()
		operatorBackOff.MaxElapsedTime = 5 * time.Minute
	}

	var operatorService *operator.Operator
	{
		c := operator.DefaultConfig()

		c.BackOff = operatorBackOff
		c.Framework = operatorFramework
		c.Informer = newInformer
		c.Logger = config.Logger
		c.TPR = newTPR

		operatorService, err = operator.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.DefaultConfig()

		versionConfig.Description = config.Description
		versionConfig.GitCommit = config.GitCommit
		versionConfig.Name = config.Name
		versionConfig.Source = config.Source

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Healthz:  healthzService,
		Operator: operatorService,
		Version:  versionService,

		bootOnce: sync.Once{},
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.Operator.Boot()
	})
}
