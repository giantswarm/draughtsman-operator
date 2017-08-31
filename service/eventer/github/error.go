package github

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notFoundError = microerror.New("not found")

// IsNotFound asserts notFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == notFoundError
}

var unexpectedStatusCode = microerror.New("unexpected status code")

// IsUnexpectedStatusCode asserts unexpectedStatusCode.
func IsUnexpectedStatusCode(err error) bool {
	return microerror.Cause(err) == unexpectedStatusCode
}
