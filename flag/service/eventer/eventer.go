package eventer

import (
	"github.com/giantswarm/draughtsman-operator/flag/service/eventer/github"
)

type Eventer struct {
	GitHub github.GitHub
	Type   string
}
