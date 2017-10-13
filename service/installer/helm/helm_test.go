package helm

import (
	"reflect"
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

// Test_Installer_Helm_chartName tests the chartName method.
func Test_Installer_Helm_chartName(t *testing.T) {
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
			expectedChartName: "giantswarm_api-chart_1.0.0-12345/api-chart",
		},
	}

	for index, test := range tests {
		i := Installer{
			registry:     test.registry,
			organisation: test.organisation,
		}

		returnedChartName := i.chartName(test.project)

		if returnedChartName != test.expectedChartName {
			t.Fatalf(
				"%v\nexpected: %#v\nreturned: %#v\n",
				index, test.expectedChartName, returnedChartName,
			)
		}
	}
}

func Test_Installer_Helm_bytesToProjects(t *testing.T) {
	testCases := []struct {
		Bytes            []byte
		ExpectedProjects []spec.Project
	}{
		{
			Bytes: []byte(`
NAME               	REVISION	UPDATED                 	STATUS  	CHART                                                       	NAMESPACE
api                	4       	Wed Aug 30 19:32:47 2017	DEPLOYED	api-chart-1.0.0-8df9e731276736f91106765073cbcbc9ac45248b    	default
cluster-service    	1       	Wed Aug 30 19:32:52 2017	DEPLOYED	cluster-service-chart-1.0.0-1de4cedf870ba17b46d775070160a...	default
etcd-operator-0-1-0	1       	Wed Aug 30 19:27:55 2017	DEPLOYED	etcd-operator-0.4.3                                         	default
			`),
			ExpectedProjects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160a",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "",
				},
			},
		},
	}

	for i, tc := range testCases {
		projects, err := bytesToProjects(tc.Bytes)
		if err != nil {
			t.Fatalf("test %d expected %#v got %#v", i+1, nil, err)
		}

		if !reflect.DeepEqual(tc.ExpectedProjects, projects) {
			t.Fatalf("test %d expected %#v got %#v", i+1, tc.ExpectedProjects, projects)
		}
	}
}
