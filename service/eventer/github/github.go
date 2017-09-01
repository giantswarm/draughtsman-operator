package github

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	eventerspec "github.com/giantswarm/draughtsman-operator/service/eventer/spec"
	httpspec "github.com/giantswarm/draughtsman-operator/service/http"
)

var (
	// GithubEventerType is an Eventer that uses Github Deployment Events as a backend.
	GithubEventerType eventerspec.EventerType = "GithubEventer"
)

// Config represents the configuration used to create a GitHub Eventer.
type Config struct {
	// Dependencies.
	HTTPClient httpspec.Client
	Logger     micrologger.Logger

	// Settings.
	OAuthToken   string
	Organisation string
}

// DefaultConfig provides a default configuration to create a new GitHub
// Eventer by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		HTTPClient: nil,
		Logger:     nil,

		// Settings.
		OAuthToken:   "",
		Organisation: "",
	}
}

// Eventer is an implementation of the Eventer interface,
// that uses GitHub Deployment Events as a backend.
type Eventer struct {
	// Dependencies.
	client httpspec.Client
	logger micrologger.Logger

	// Settings.
	oauthToken   string
	organisation string
}

// New creates a new configured GitHub Eventer.
func New(config Config) (*Eventer, error) {
	// Dependencies.
	if config.HTTPClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.HTTPClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	if config.OAuthToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.OAuthToken token must not be empty")
	}
	if config.Organisation == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Organisation must not be empty")
	}

	eventer := &Eventer{
		// Dependencies.
		client: config.HTTPClient,
		logger: config.Logger,

		// Settings.
		oauthToken:   config.OAuthToken,
		organisation: config.Organisation,
	}

	return eventer, nil
}

func (e *Eventer) SetFailedStatus(event eventerspec.DeploymentEvent) error {
	return e.postDeploymentStatus(event.Name, event.ID, failureState)
}

func (e *Eventer) SetSuccessStatus(event eventerspec.DeploymentEvent) error {
	return e.postDeploymentStatus(event.Name, event.ID, successState)
}
