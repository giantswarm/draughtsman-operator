package project

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var desiredProjects []Project

	for _, p := range customObject.Spec.Projects {
		desiredProjects = append(desiredProjects, Project{ID: p.ID, Name: p.Name, Ref: p.Ref})
	}

	return desiredProjects, nil
}
