package project

import (
	"fmt"

	"github.com/giantswarm/draughtsmantpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

const (
	// Name is the identifier of the resource.
	Name = "project"
)

// Config represents the configuration used to create a new project resource.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new project
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,
	}
}

// Resource implements the project resource.
type Resource struct {
	// Dependencies.
	logger micrologger.Logger
}

// New creates a new configured project resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	newService := &Resource{
		// Dependencies.
		logger: config.Logger.With(
			"resource", Name,
		),
	}

	return newService, nil
}

func (r *Resource) GetCurrentState(obj interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get current state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", fmt.Sprintf("found k8s state: %#v", nil))

	return nil, nil
}

func (r *Resource) GetDesiredState(obj interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get desired state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", fmt.Sprintf("found desired state: %#v", nil))

	return nil, nil
}

func (r *Resource) GetCreateState(obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get create state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", fmt.Sprintf("found create state: %#v", nil))

	return nil, nil
}

func (r *Resource) GetDeleteState(obj, currentState, desiredState interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get delete state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", fmt.Sprintf("found delete state: %#v", nil))

	return nil, nil
}

func (r *Resource) GetUpdateState(obj, currentState, desiredState interface{}) (interface{}, interface{}, interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get delete state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", fmt.Sprintf("found delete state: %#v", nil))

	return nil, nil, nil, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(obj, createState interface{}) error {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("debug", "process create state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", "processed create state")

	return nil
}

func (r *Resource) ProcessDeleteState(obj, deleteState interface{}) error {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("debug", "process delete state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", "processed delete state")

	return nil
}

func (r *Resource) ProcessUpdateState(obj, updateState interface{}) error {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("debug", "process update state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", "processed update state")

	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func toCustomObject(v interface{}) (draughtsmantpr.CustomObject, error) {
	customObjectPointer, ok := v.(*draughtsmantpr.CustomObject)
	if !ok {
		return draughtsmantpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &draughtsmantpr.CustomObject{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}
