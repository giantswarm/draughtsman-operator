package installer

import (
	"github.com/giantswarm/draughtsman-operator/flag/service/installer/helm"
)

type Installer struct {
	Helm helm.Helm
	Type string
}
