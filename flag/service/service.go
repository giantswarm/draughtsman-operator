package service

import (
	"github.com/giantswarm/draughtsman-operator/flag/service/configurer"
	"github.com/giantswarm/draughtsman-operator/flag/service/helm"
	"github.com/giantswarm/draughtsman-operator/flag/service/kubernetes"
)

type Service struct {
	Configurer configurer.Configurer
	Helm       helm.Helm
	Kubernetes kubernetes.Kubernetes
}
