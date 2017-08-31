package configurer

import (
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/draughtsman-operator/flag"
	"github.com/giantswarm/draughtsman-operator/service/configurer/configmap"
	"github.com/giantswarm/draughtsman-operator/service/configurer/file"
	"github.com/giantswarm/draughtsman-operator/service/configurer/secret"
	"github.com/giantswarm/draughtsman-operator/service/configurer/spec"
)

// Config represents the configuration used to create a Configurer.
type Config struct {
	// Dependencies.
	FileSystem afero.Fs
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper

	Type spec.ConfigurerType
}

// DefaultConfig provides a default configuration to create a new Configurer
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		FileSystem: nil,
		K8sClient:  nil,
		Logger:     nil,

		// Settings.
		Flag:  nil,
		Viper: nil,
	}
}

// New creates a new configured Configurer.
func New(config Config) (spec.Configurer, error) {
	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "viper must not be empty")
	}

	var err error

	var newConfigurer spec.Configurer
	switch config.Type {
	case configmap.ConfigurerType:
		configmapConfig := configmap.DefaultConfig()

		configmapConfig.K8sClient = config.K8sClient
		configmapConfig.Logger = config.Logger

		configmapConfig.Key = config.Viper.GetString(config.Flag.Service.Configurer.ConfigMap.Key)
		configmapConfig.Name = config.Viper.GetString(config.Flag.Service.Configurer.ConfigMap.Name)
		configmapConfig.Namespace = config.Viper.GetString(config.Flag.Service.Configurer.ConfigMap.Namespace)

		newConfigurer, err = configmap.New(configmapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	case file.ConfigurerType:
		fileConfig := file.DefaultConfig()

		fileConfig.FileSystem = config.FileSystem
		fileConfig.Logger = config.Logger

		fileConfig.Path = config.Viper.GetString(config.Flag.Service.Configurer.File.Path)

		newConfigurer, err = file.New(fileConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	case secret.ConfigurerType:
		secretConfig := secret.DefaultConfig()

		secretConfig.K8sClient = config.K8sClient
		secretConfig.Logger = config.Logger

		secretConfig.Key = config.Viper.GetString(config.Flag.Service.Configurer.Secret.Key)
		secretConfig.Name = config.Viper.GetString(config.Flag.Service.Configurer.Secret.Name)
		secretConfig.Namespace = config.Viper.GetString(config.Flag.Service.Configurer.Secret.Namespace)

		newConfigurer, err = secret.New(secretConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	default:
		return nil, microerror.Maskf(invalidConfigError, "configurer type not implemented")
	}

	return newConfigurer, nil
}
