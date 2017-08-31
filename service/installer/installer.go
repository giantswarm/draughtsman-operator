package installer

import (
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/draughtsman-operator/flag"
	"github.com/giantswarm/draughtsman-operator/service/configurer"
	configurerspec "github.com/giantswarm/draughtsman-operator/service/configurer/spec"
	"github.com/giantswarm/draughtsman-operator/service/installer/helm"
	"github.com/giantswarm/draughtsman-operator/service/installer/spec"
)

// Config represents the configuration used to create an Installer.
type Config struct {
	// Dependencies.
	FileSystem afero.Fs
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper

	Type spec.InstallerType
}

// DefaultConfig provides a default configuration to create a new Installer
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		FileSystem: nil,
		K8sClient:  nil,
		Logger:     nil,

		// Settings.
		Flag:  nil,
		Type:  "",
		Viper: nil,
	}
}

// New creates a new configured Installer.
func New(config Config) (spec.Installer, error) {
	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Type == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Type must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var configurerServices []configurerspec.Configurer
	types := strings.Split(config.Viper.GetString(config.Flag.Service.Configurer.Types), ",")
	for _, t := range types {
		configurerConfig := configurer.DefaultConfig()

		configurerConfig.FileSystem = config.FileSystem
		configurerConfig.K8sClient = config.K8sClient
		configurerConfig.Logger = config.Logger

		configurerConfig.Flag = config.Flag
		configurerConfig.Type = configurerspec.ConfigurerType(t)
		configurerConfig.Viper = config.Viper

		configurerService, err := configurer.New(configurerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configurerServices = append(configurerServices, configurerService)
	}

	var newInstaller spec.Installer
	switch config.Type {
	case helm.HelmInstallerType:
		helmConfig := helm.DefaultConfig()

		helmConfig.Configurers = configurerServices
		helmConfig.FileSystem = config.FileSystem
		helmConfig.Logger = config.Logger

		helmConfig.HelmBinaryPath = config.Viper.GetString(config.Flag.Service.Installer.Helm.HelmBinaryPath)
		helmConfig.Organisation = config.Viper.GetString(config.Flag.Service.Installer.Helm.Organisation)
		helmConfig.Password = config.Viper.GetString(config.Flag.Service.Installer.Helm.Password)
		helmConfig.Registry = config.Viper.GetString(config.Flag.Service.Installer.Helm.Registry)
		helmConfig.Username = config.Viper.GetString(config.Flag.Service.Installer.Helm.Username)

		newInstaller, err = helm.New(helmConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	default:
		return nil, microerror.Maskf(invalidConfigError, "installer type not implemented")
	}

	return newInstaller, nil
}
