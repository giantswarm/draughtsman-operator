package service

import (
	"github.com/giantswarm/draughtsman-operator/flag/service/kubernetes"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
}
