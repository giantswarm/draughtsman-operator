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

func Test_Installer_Helm_getProjectFromList(t *testing.T) {
	testCases := []struct {
		Projects     []spec.Project
		Project      spec.Project
		ErrorMatcher func(error) bool
		Expected     spec.Project
	}{
		{
			Projects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160abcbc9ac45248b",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "xde5cedfg777a17846d77h088n60a0cbc9ac4594fb",
				},
			},
			Project: spec.Project{
				Name: "api",
				Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
			},
			ErrorMatcher: nil,
			Expected: spec.Project{
				Name: "api",
				Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
			},
		},

		{
			Projects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160abcbc9ac45248b",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "xde5cedfg777a17846d77h088n60a0cbc9ac4594fb",
				},
			},
			Project: spec.Project{
				Name: "api",
				Ref:  "8df9e731276736f91106765073cbcb",
			},
			ErrorMatcher: nil,
			Expected: spec.Project{
				Name: "api",
				Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
			},
		},

		{
			Projects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160abcbc9ac45248b",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "xde5cedfg777a17846d77h088n60a0cbc9ac4594fb",
				},
			},
			Project: spec.Project{
				Name: "api",
				Ref:  "8df9e7312767",
			},
			ErrorMatcher: nil,
			Expected: spec.Project{
				Name: "api",
				Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
			},
		},

		{
			Projects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160abcbc9ac45248b",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "xde5cedfg777a17846d77h088n60a0cbc9ac4594fb",
				},
			},
			Project: spec.Project{
				Name: "api",
				Ref:  "8df9",
			},
			ErrorMatcher: nil,
			Expected: spec.Project{
				Name: "api",
				Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
			},
		},

		{
			Projects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160abcbc9ac45248b",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "xde5cedfg777a17846d77h088n60a0cbc9ac4594fb",
				},
			},
			Project: spec.Project{
				Name: "api-service",
				Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
			},
			ErrorMatcher: IsNotFound,
			Expected:     spec.Project{},
		},

		{
			Projects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160abcbc9ac45248b",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "xde5cedfg777a17846d77h088n60a0cbc9ac4594fb",
				},
			},
			Project: spec.Project{
				Name: "api-service",
				Ref:  "8df9e731276736f91106765073cbcb",
			},
			ErrorMatcher: IsNotFound,
			Expected:     spec.Project{},
		},

		{
			Projects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160abcbc9ac45248b",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "xde5cedfg777a17846d77h088n60a0cbc9ac4594fb",
				},
			},
			Project: spec.Project{
				Name: "api-service",
				Ref:  "8df9e7312767",
			},
			ErrorMatcher: IsNotFound,
			Expected:     spec.Project{},
		},

		{
			Projects: []spec.Project{
				{
					Name: "api",
					Ref:  "8df9e731276736f91106765073cbcbc9ac45248b",
				},
				{
					Name: "cluster-service",
					Ref:  "1de4cedf870ba17b46d775070160abcbc9ac45248b",
				},
				{
					Name: "etcd-operator-0-1-0",
					Ref:  "xde5cedfg777a17846d77h088n60a0cbc9ac4594fb",
				},
			},
			Project: spec.Project{
				Name: "api-service",
				Ref:  "8df9",
			},
			ErrorMatcher: IsNotFound,
			Expected:     spec.Project{},
		},
	}

	for i, tc := range testCases {
		foundProject, err := getProjectFromList(tc.Projects, tc.Project)
		if err != nil {
			if tc.ErrorMatcher != nil {
				if !tc.ErrorMatcher(err) {
					t.Fatalf("test %d expected %#v got %#v", i+1, true, false)
				}
			} else {
				t.Fatalf("test %d expected %#v got %#v", i+1, nil, err)
			}
		}
		if !reflect.DeepEqual(tc.Expected, foundProject) {
			t.Fatalf("test %d expected %#v got %#v", i+1, tc.Expected, foundProject)
		}
	}
}
