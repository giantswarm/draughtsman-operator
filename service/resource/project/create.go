package project

import (
	"context"
	"strconv"

	"github.com/giantswarm/microerror"

	eventerspec "github.com/giantswarm/draughtsman-operator/service/eventer/spec"
	installerspec "github.com/giantswarm/draughtsman-operator/service/installer/spec"
	notifierspec "github.com/giantswarm/draughtsman-operator/service/notifier/spec"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	projectsToCreate, err := toProjects(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(projectsToCreate) != 0 {
		r.logger.Log("debug", "creating projects in the Kubernetes cluster")

		for _, p := range projectsToCreate {
			ID, err := strconv.Atoi(p.ID)
			if err != nil {
				return microerror.Mask(err)
			}

			instErr := r.installer.Install(installerspec.Project{Name: p.Name, Ref: p.Ref})
			if instErr != nil {
				evenErr := r.eventer.SetFailedStatus(eventerspec.DeploymentEvent{ID: ID, Name: p.Name, Sha: p.Ref})
				if evenErr != nil {
					return microerror.Mask(evenErr)
				}
				notiErr := r.notifier.Failed(notifierspec.Project{ID: p.ID, Name: p.Name, Ref: p.Ref}, instErr.Error())
				if notiErr != nil {
					return microerror.Mask(notiErr)
				}

				return microerror.Mask(instErr)
			}

			evenErr := r.eventer.SetSuccessStatus(eventerspec.DeploymentEvent{ID: ID, Name: p.Name, Sha: p.Ref})
			if evenErr != nil {
				return microerror.Mask(evenErr)
			}
			notiErr := r.notifier.Success(notifierspec.Project{ID: p.ID, Name: p.Name, Ref: p.Ref})
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

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
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
