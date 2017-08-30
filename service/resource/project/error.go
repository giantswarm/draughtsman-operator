package project

import (
	"github.com/giantswarm/microerror"
)

var helmError = microerror.New("helm error")

// IsHelm asserts helmError.
func IsHelm(err error) bool {
	return microerror.Cause(err) == helmError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var wrongTypeError = microerror.New("wrong type")

// IsWrongType asserts wrongTypeError.
func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
