package project

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	r.logger.Log("TODO", "implement ApplyDeleteChange")
	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
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
