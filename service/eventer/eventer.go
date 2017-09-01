package eventer

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/draughtsman-operator/flag"
	"github.com/giantswarm/draughtsman-operator/service/eventer/github"
	"github.com/giantswarm/draughtsman-operator/service/eventer/spec"
	eventerspec "github.com/giantswarm/draughtsman-operator/service/eventer/spec"
	httpspec "github.com/giantswarm/draughtsman-operator/service/http"
)

// Config represents the configuration used to create an Eventer.
type Config struct {
	// Dependencies.
	HTTPClient httpspec.Client
	Logger     micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper
}

// DefaultConfig provides a default configuration to create a new Eventer
// service by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		HTTPClient: nil,
		Logger:     nil,

		// Settings.
		Flag:  nil,
		Viper: nil,
	}
}

// New creates a new configured Eventer.
func New(config Config) (spec.Eventer, error) {
	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var newEventer spec.Eventer
	t := eventerspec.EventerType(config.Viper.GetString(config.Flag.Service.Eventer.Type))
	switch t {
	case github.GithubEventerType:
		githubConfig := github.DefaultConfig()

		githubConfig.HTTPClient = config.HTTPClient
		githubConfig.Logger = config.Logger

		githubConfig.OAuthToken = config.Viper.GetString(config.Flag.Service.Eventer.GitHub.OAuthToken)
		githubConfig.Organisation = config.Viper.GetString(config.Flag.Service.Eventer.GitHub.Organisation)

		newEventer, err = github.New(githubConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	default:
		return nil, microerror.Maskf(invalidConfigError, "eventer type '%s' not implemented", t)
	}

	return newEventer, nil
}
