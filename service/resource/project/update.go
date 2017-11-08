package project

import (
	"context"
	"strconv"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"

	eventerspec "github.com/giantswarm/draughtsman-operator/service/eventer/spec"
	installerspec "github.com/giantswarm/draughtsman-operator/service/installer/spec"
	notifierspec "github.com/giantswarm/draughtsman-operator/service/notifier/spec"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	projectsToUpdate, err := toProjects(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(projectsToUpdate) != 0 {
		r.logger.Log("debug", "updating projects in the Kubernetes cluster")

		for _, p := range projectsToUpdate {
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

		r.logger.Log("debug", "updated projects in the Kubernetes cluster")
	} else {
		r.logger.Log("debug", "the projects are already up to date in the Kubernetes cluster")
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentProjects, err := toProjects(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredProjects, err := toProjects(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var projectsToUpdate []Project

	for _, desiredProject := range desiredProjects {
		if !existsProjectByName(currentProjects, desiredProject.Name) {
			continue
		}

		currentProject, err := getProjectByName(currentProjects, desiredProject.Name)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		// NOTE that we need to deal with eventually incomplete sha/ref information
		// given in the list of current projects. This is due to certain helm
		// limitations.
		if !strings.HasPrefix(desiredProject.Ref, currentProject.Ref) {
			projectsToUpdate = append(projectsToUpdate, desiredProject)
		}
	}

	return projectsToUpdate, nil
}
