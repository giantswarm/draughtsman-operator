package helm

import (
	"testing"

	"github.com/giantswarm/draughtsman-operator/service/installer/spec"
)

// Test_Installer_Helm_versionedChartName tests the versionedChartName method.
func Test_Installer_Helm_versionedChartName(t *testing.T) {
	tests := []struct {
		registry          string
		organisation      string
		project           spec.Project
		expectedChartName string
	}{
		{
			registry:     "quay.io",
			organisation: "giantswarm",
			project: spec.Project{
				Name: "api",
				Ref:  "12345",
			},
			expectedChartName: "quay.io/giantswarm/api-chart@1.0.0-12345",
		},
	}

	for index, test := range tests {
		i := Installer{
			registry:     test.registry,
			organisation: test.organisation,
		}

		returnedChartName := i.versionedChartName(test.project)

		if returnedChartName != test.expectedChartName {
			t.Fatalf(
				"%v\nexpected: %#v\nreturned: %#v\n",
				index, test.expectedChartName, returnedChartName,
			)
		}
	}
}

// Test_Installer_Helm_tarballName tests the tarballName method.
func Test_Installer_Helm_tarballName(t *testing.T) {
	tests := []struct {
		registry            string
		organisation        string
		project             spec.Project
		expectedTarballName string
	}{
		{
			registry:     "quay.io",
			organisation: "giantswarm",
			project: spec.Project{
				Name: "api",
				Ref:  "12345",
			},
			expectedTarballName: "giantswarm_api-chart_1.0.0-12345.tar.gz",
		},
	}

	for index, test := range tests {
		i := Installer{
			registry:     test.registry,
			organisation: test.organisation,
		}

		returnedTarballName := i.tarballName(test.project)

		if returnedTarballName != test.expectedTarballName {
			t.Fatalf(
				"%v\nexpected: %#v\nreturned: %#v\n",
				index, test.expectedTarballName, returnedTarballName,
			)
		}
	}
}
