package secret

import (
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/draughtsman-operator/service/configurer/spec"
)

// ConfigurerType is the kind of a Configurer that is backed by a Kubernetes
// Secret.
var ConfigurerType spec.ConfigurerType = "SecretConfigurer"

// Config represents the configuration used to create a Secret Configurer.
type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Settings.

	// Key is the key to reference the values data in the secret.
	Key       string
	Name      string
	Namespace string
}

// DefaultConfig provides a default configuration to create a new Secret
// Configurer by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient: nil,
		Logger:    nil,

		// Settings.
		Key:       "",
		Name:      "",
		Namespace: "",
	}
}

// SecretConfigurer is an implementation of the Configurer interface, that uses
// a Kubernetes Secret to hold configuration.
type SecretConfigurer struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	// Settings.
	key       string
	name      string
	namespace string
}

// New creates a new configured Secret Configurer.
func New(config Config) (*SecretConfigurer, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	if config.Key == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Key must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Name must not be empty")
	}
	if config.Namespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Namespace must not be empty")
	}

	configurer := &SecretConfigurer{
		// Dependencies.
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		// Settings.
		key:       config.Key,
		name:      config.Name,
		namespace: config.Namespace,
	}

	return configurer, nil
}

func (c *SecretConfigurer) Type() spec.ConfigurerType {
	return ConfigurerType
}

func (c *SecretConfigurer) Values() (string, error) {
	defer updateSecretMetrics(time.Now())

	c.logger.Log("debug", "fetching configuration from secret", "name", c.name, "namespace", c.namespace)

	s, err := c.k8sClient.CoreV1().Secrets(c.namespace).Get(c.name, v1.GetOptions{})
	if err != nil {
		return "", microerror.Mask(err)
	}

	var valuesData string
	{
		b, ok := s.Data[c.key]
		if !ok {
			return "", microerror.Maskf(keyMissingError, "key '%s' not found in secret", c.key)
		}
		valuesData = string(b)
	}

	return valuesData, nil
}
