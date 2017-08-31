package service

import (
	"github.com/giantswarm/draughtsman-operator/flag/service/configurer"
	"github.com/giantswarm/draughtsman-operator/flag/service/eventer"
	"github.com/giantswarm/draughtsman-operator/flag/service/installer"
	"github.com/giantswarm/draughtsman-operator/flag/service/kubernetes"
	"github.com/giantswarm/draughtsman-operator/flag/service/notifier"
	"github.com/giantswarm/draughtsman/flag/service/httpclient"
)

type Service struct {
	Configurer configurer.Configurer
	Eventer    eventer.Eventer
	HTTPClient httpclient.HTTPClient
	Installer  installer.Installer
	Kubernetes kubernetes.Kubernetes
	Notifier   notifier.Notifier
}
