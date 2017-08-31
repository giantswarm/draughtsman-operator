package eventer

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/draughtsman-operator/service/eventer/github"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var NotFoundError = microerror.New("not found")

// IsNotFound asserts not found errors of eventer implementations.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == NotFoundError || github.IsNotFound(err)
}
