package service

import (
	"github.com/giantswarm/draughtsman-operator/flag/service/configurer"
	"github.com/giantswarm/draughtsman-operator/flag/service/installer"
	"github.com/giantswarm/draughtsman-operator/flag/service/kubernetes"
)

type Service struct {
	Configurer configurer.Configurer
	Installer  installer.Installer
	Kubernetes kubernetes.Kubernetes
}
