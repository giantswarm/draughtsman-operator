package draughtsmantpr

import "github.com/giantswarm/draughtsmantpr/spec"

type Spec struct {
	Projects []spec.Project `json:"projects" yaml:"projects"`
}
