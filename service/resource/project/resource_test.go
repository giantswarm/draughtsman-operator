package project

import (
	"testing"
)

// Test_Resource_Project_versionedChartName tests the versionedChartName method.
func Test_Resource_Project_versionedChartName(t *testing.T) {
	tests := []struct {
		registry          string
		organisation      string
		project           Project
		expectedChartName string
	}{
		{
			registry:     "quay.io",
			organisation: "giantswarm",
			project: Project{
				Name: "api",
				Ref:  "12345",
			},
			expectedChartName: "quay.io/giantswarm/api-chart@1.0.0-12345",
		},
	}

	for index, test := range tests {
		r := Resource{
			registry:     test.registry,
			organisation: test.organisation,
		}

		returnedChartName := r.versionedChartName(test.project)

		if returnedChartName != test.expectedChartName {
			t.Fatalf(
				"%v\nexpected: %#v\nreturned: %#v\n",
				index, test.expectedChartName, returnedChartName,
			)
		}
	}
}

// Test_Resource_Project_tarballName tests the tarballName method.
func Test_Resource_Project_tarballName(t *testing.T) {
	tests := []struct {
		registry            string
		organisation        string
		project             Project
		expectedTarballName string
	}{
		{
			registry:     "quay.io",
			organisation: "giantswarm",
			project: Project{
				Name: "api",
				Ref:  "12345",
			},
			expectedTarballName: "giantswarm_api-chart_1.0.0-12345.tar.gz",
		},
	}

	for index, test := range tests {
		r := Resource{
			registry:     test.registry,
			organisation: test.organisation,
		}

		returnedTarballName := r.tarballName(test.project)

		if returnedTarballName != test.expectedTarballName {
			t.Fatalf(
				"%v\nexpected: %#v\nreturned: %#v\n",
				index, test.expectedTarballName, returnedTarballName,
			)
		}
	}
}
