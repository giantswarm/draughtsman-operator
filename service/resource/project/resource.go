package project

import (
	"strings"

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

func (r *Resource) GetCurrentState(obj interface{}) (interface{}, error) {
	var currentProjects []Project

	{
		list, err := r.installer.List()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		currentProjects = installerProjectsToProjects(list)
	}

	return currentProjects, nil
}

func (r *Resource) GetDesiredState(obj interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var desiredProjects []Project

	for _, p := range customObject.Spec.Projects {
		desiredProjects = append(desiredProjects, Project{Name: p.Name, Ref: p.Ref})
	}

	return desiredProjects, nil
}

func (r *Resource) GetCreateState(obj, currentState, desiredState interface{}) (interface{}, error) {
	currentProjects, err := toProjects(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredProjects, err := toProjects(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var projectsToCreate []Project

	for _, desiredProject := range desiredProjects {
		if !existsProjectByName(currentProjects, desiredProject.Name) {
			projectsToCreate = append(projectsToCreate, desiredProject)
		}
	}

	return projectsToCreate, nil
}

func (r *Resource) GetDeleteState(obj, currentState, desiredState interface{}) (interface{}, error) {
	currentProjects, err := toProjects(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredProjects, err := toProjects(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var projectsToDelete []Project

	for _, currentProject := range currentProjects {
		if existsProjectByName(desiredProjects, currentProject.Name) {
			projectsToDelete = append(projectsToDelete, currentProject)
		}
	}

	return projectsToDelete, nil
}

func (r *Resource) GetUpdateState(obj, currentState, desiredState interface{}) (interface{}, interface{}, interface{}, error) {
	currentProjects, err := toProjects(currentState)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}
	desiredProjects, err := toProjects(desiredState)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}

	var projectsToUpdate []Project

	for _, desiredProject := range desiredProjects {
		if !existsProjectByName(currentProjects, desiredProject.Name) {
			continue
		}

		currentProject, err := getProjectByName(currentProjects, desiredProject.Name)
		if err != nil {
			return nil, nil, nil, microerror.Mask(err)
		}

		// NOTE that we need to deal with eventually incomplete sha/ref information
		// given in the list of current projects. This is due to certain helm
		// limitations.
		if !strings.HasPrefix(desiredProject.Ref, currentProject.Ref) {
			projectsToUpdate = append(projectsToUpdate, desiredProject)
		}
	}

	return []Project{}, []Project{}, projectsToUpdate, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(obj, createState interface{}) error {
	projectsToCreate, err := toProjects(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(projectsToCreate) != 0 {
		r.logger.Log("debug", "creating projects in the Kubernetes cluster")

		for _, p := range projectsToCreate {
			instErr := r.installer.Install(installerspec.Project{Name: p.Name, Ref: p.Ref})
			if instErr != nil {
				evenErr := r.eventer.SetFailedStatus(eventerspec.DeploymentEvent{ID: 0, Name: p.Name, Sha: p.Ref})
				if evenErr != nil {
					return microerror.Mask(evenErr)
				}
				notiErr := r.notifier.Failed(notifierspec.Project{Name: p.Name, Ref: p.Ref}, instErr.Error())
				if notiErr != nil {
					return microerror.Mask(notiErr)
				}

				return microerror.Mask(instErr)
			}

			evenErr := r.eventer.SetSuccessStatus(eventerspec.DeploymentEvent{ID: 0, Name: p.Name, Sha: p.Ref})
			if evenErr != nil {
				return microerror.Mask(evenErr)
			}
			notiErr := r.notifier.Success(notifierspec.Project{Name: p.Name, Ref: p.Ref})
			if notiErr != nil {
				return microerror.Mask(notiErr)
			}
		}

		r.logger.Log("debug", "created projects in the Kubernetes cluster")
	} else {
		r.logger.Log("debug", "the projects are already created in the Kubernetes cluster")
	}

	return nil
}

func (r *Resource) ProcessDeleteState(obj, deleteState interface{}) error {
	r.logger.Log("TODO", "implement ProcessDeleteState")
	return nil
}

func (r *Resource) ProcessUpdateState(obj, updateState interface{}) error {
	projectsToUpdate, err := toProjects(updateState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(projectsToUpdate) != 0 {
		r.logger.Log("debug", "updating projects in the Kubernetes cluster")

		for _, p := range projectsToUpdate {
			instErr := r.installer.Install(installerspec.Project{Name: p.Name, Ref: p.Ref})
			if instErr != nil {
				evenErr := r.eventer.SetFailedStatus(eventerspec.DeploymentEvent{ID: 0, Name: p.Name, Sha: p.Ref})
				if evenErr != nil {
					return microerror.Mask(evenErr)
				}
				notiErr := r.notifier.Failed(notifierspec.Project{Name: p.Name, Ref: p.Ref}, instErr.Error())
				if notiErr != nil {
					return microerror.Mask(notiErr)
				}

				return microerror.Mask(instErr)
			}

			evenErr := r.eventer.SetSuccessStatus(eventerspec.DeploymentEvent{ID: 0, Name: p.Name, Sha: p.Ref})
			if evenErr != nil {
				return microerror.Mask(evenErr)
			}
			notiErr := r.notifier.Success(notifierspec.Project{Name: p.Name, Ref: p.Ref})
			if notiErr != nil {
				return microerror.Mask(notiErr)
			}
		}

		r.logger.Log("debug", "updated projects in the Kubernetes cluster")
	} else {
		r.logger.Log("debug", "the projects are already up to date in the Kubernetes cluster")
	}

	return nil
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
