package project

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
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
