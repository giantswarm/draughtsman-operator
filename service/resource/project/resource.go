package project

import (
	"github.com/giantswarm/draughtsmantpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"

	eventerspec "github.com/giantswarm/draughtsman-operator/service/eventer/spec"
	installerspec "github.com/giantswarm/draughtsman-operator/service/installer/spec"
	notifierspec "github.com/giantswarm/draughtsman-operator/service/notifier/spec"
)

const (
	// Name is the identifier of the resource.
	Name = "project"
)

// Config represents the configuration used to create a new project resource.
type Config struct {
	// Dependencies.
	Eventer   eventerspec.Eventer
	Installer installerspec.Installer
	Logger    micrologger.Logger
	Notifier  notifierspec.Notifier
}

// DefaultConfig provides a default configuration to create a new project
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Eventer:   nil,
		Installer: nil,
		Logger:    nil,
		Notifier:  nil,
	}
}

// Resource implements the project resource.
type Resource struct {
	// Dependencies.
	eventer   eventerspec.Eventer
	installer installerspec.Installer
	logger    micrologger.Logger
	notifier  notifierspec.Notifier
}

// New creates a new configured project resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Eventer == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Eventer must not be empty")
	}
	if config.Installer == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Installer must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Notifier == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Notifier must not be empty")
	}

	newResource := &Resource{
		// Dependencies.
		eventer:   config.Eventer,
		installer: config.Installer,
		logger: config.Logger.With(
			"resource", Name,
		),
		notifier: config.Notifier,
	}

	return newResource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func existsProjectByName(list []Project, name string) bool {
	for _, l := range list {
		if l.Name == name {
			return true
		}
	}

	return false
}

func getProjectByName(list []Project, name string) (Project, error) {
	for _, l := range list {
		if l.Name == name {
			return l, nil
		}
	}

	return Project{}, microerror.Maskf(notFoundError, name)
}

func installerProjectsToProjects(installerList []installerspec.Project) []Project {
	var list []Project

	for _, p := range installerList {
		list = append(list, Project{Name: p.Name, Ref: p.Ref})
	}

	return list
}

func toCustomObject(v interface{}) (draughtsmantpr.CustomObject, error) {
	customObjectPointer, ok := v.(*draughtsmantpr.CustomObject)
	if !ok {
		return draughtsmantpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &draughtsmantpr.CustomObject{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func toProjects(v interface{}) ([]Project, error) {
	projects, ok := v.([]Project)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []Project{}, v)
	}

	return projects, nil
}
