package project

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/giantswarm/draughtsmantpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/spf13/afero"

	configurerspec "github.com/giantswarm/draughtsman-operator/service/configurer/spec"
)

const (
	// Name is the identifier of the resource.
	Name = "project"
	// tarballNameFormat is the format for the name of the chart tarball.
	tarballNameFormat = "%v_%v-chart_1.0.0-%v.tar.gz"
	// versionedChartFormat is the format the CNR registry uses to address
	// charts. For example, we use this to address that chart to pull.
	versionedChartFormat = "%v/%v/%v-chart@1.0.0-%v"
)

// Config represents the configuration used to create a new project resource.
type Config struct {
	// Dependencies.
	Configurers []configurerspec.Configurer
	FileSystem  afero.Fs
	Logger      micrologger.Logger

	// Settings.
	HelmBinaryPath string
	Organisation   string
	Password       string
	Registry       string
	Username       string
}

// DefaultConfig provides a default configuration to create a new project
// resource by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Configurers: nil,
		FileSystem:  afero.NewMemMapFs(),
		Logger:      nil,

		// Settings.
		HelmBinaryPath: "",
		Organisation:   "",
		Password:       "",
		Registry:       "",
		Username:       "",
	}
}

// Resource implements the project resource.
type Resource struct {
	// Dependencies.
	configurers []configurerspec.Configurer
	fileSystem  afero.Fs
	logger      micrologger.Logger

	// Settings.
	helmBinaryPath string
	organisation   string
	password       string
	registry       string
	username       string
}

// New creates a new configured project resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Configurers == nil {
		return nil, microerror.Maskf(invalidConfigError, "configurers must not be empty")
	}
	if config.FileSystem == nil {
		return nil, microerror.Maskf(invalidConfigError, "file system must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	// Settings.
	if config.HelmBinaryPath == "" {
		return nil, microerror.Maskf(invalidConfigError, "helm binary path must not be empty")
	}
	if config.Organisation == "" {
		return nil, microerror.Maskf(invalidConfigError, "organisation must not be empty")
	}
	if config.Password == "" {
		return nil, microerror.Maskf(invalidConfigError, "password must not be empty")
	}
	if config.Registry == "" {
		return nil, microerror.Maskf(invalidConfigError, "registry must not be empty")
	}
	if config.Username == "" {
		return nil, microerror.Maskf(invalidConfigError, "username must not be empty")
	}

	exists, err := afero.Exists(config.FileSystem, config.HelmBinaryPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if !exists {
		return nil, microerror.Maskf(invalidConfigError, "helm binary path '%s' does not exist", config.HelmBinaryPath)
	}

	newResource := &Resource{
		// Dependencies.
		configurers: config.Configurers,
		fileSystem:  config.FileSystem,
		logger: config.Logger.With(
			"resource", Name,
		),

		// Settings.
		helmBinaryPath: config.HelmBinaryPath,
		organisation:   config.Organisation,
		password:       config.Password,
		registry:       config.Registry,
		username:       config.Username,
	}

	err = newResource.login()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return newResource, nil
}

func (r *Resource) GetCurrentState(obj interface{}) (interface{}, error) {
	r.logger.Log("debug", "get current state")

	var currentProjects []Project
	{
		b, err := r.runHelmCommand("list", "list")
		if err != nil {
			return nil, microerror.Mask(err)
		}
		fmt.Printf("helm list: \n\n%s\n", b)

		// TODO parse helm list output into []Project
		// TODO filter all not being in customObject.Spec.Projects
	}

	r.logger.Log("debug", fmt.Sprintf("found k8s state: %#v", nil))

	return currentProjects, nil
}

func (r *Resource) GetDesiredState(obj interface{}) (interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get desired state")

	var desiredProjects []Project

	for _, p := range customObject.Spec.Projects {
		desiredProjects = append(desiredProjects, Project{Name: p.Name, Ref: p.Ref})
	}

	r.logger.Log("debug", fmt.Sprintf("found desired state: %#v", desiredProjects))

	return desiredProjects, nil
}

func (r *Resource) GetCreateState(obj, currentState, desiredState interface{}) (interface{}, error) {
	currentProjects, err := toProjects(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredProjects, err := toProjects(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get create state")

	var projectsToCreate []Project

	for _, desiredProject := range desiredProjects {
		if !containsProject(currentProjects, desiredProject) {
			projectsToCreate = append(projectsToCreate, desiredProject)
		}
	}

	r.logger.Log("debug", fmt.Sprintf("found create state: %#v", projectsToCreate))

	return projectsToCreate, nil
}

func (r *Resource) GetDeleteState(obj, currentState, desiredState interface{}) (interface{}, error) {
	currentProjects, err := toProjects(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredProjects, err := toProjects(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get delete state")

	var projectsToDelete []Project

	for _, currentProject := range currentProjects {
		if containsProject(desiredProjects, currentProject) {
			projectsToDelete = append(projectsToDelete, currentProject)
		}
	}

	r.logger.Log("debug", fmt.Sprintf("found delete state: %#v", projectsToDelete))

	return projectsToDelete, nil
}

func (r *Resource) GetUpdateState(obj, currentState, desiredState interface{}) (interface{}, interface{}, interface{}, error) {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return nil, nil, nil, microerror.Mask(err)
	}

	r.logger.Log("debug", "get delete state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", fmt.Sprintf("found delete state: %#v", nil))

	return nil, nil, nil, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) ProcessCreateState(obj, createState interface{}) error {
	projectsToCreate, err := toProjects(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("debug", "process create state")

	if projectsToCreate != nil {
		r.logger.Log("debug", "deploying the projects to the Kubernetes cluster")

		for _, p := range projectsToCreate {
			err := r.ensure(p)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		r.logger.Log("debug", "deployed the projects to the Kubernetes cluster")
	} else {
		r.logger.Log("debug", "the projects do already exist in the Kubernetes cluster")
	}

	r.logger.Log("debug", "processed create state")

	return nil
}

func (r *Resource) ProcessDeleteState(obj, deleteState interface{}) error {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("debug", "process delete state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", "processed delete state")

	return nil
}

func (r *Resource) ProcessUpdateState(obj, updateState interface{}) error {
	customObject, err := toCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Log("debug", "process update state")

	r.logger.Log("TODO", fmt.Sprintf("implement logic based on received custom object: %#v", customObject))

	r.logger.Log("debug", "processed update state")

	return nil
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func (r *Resource) ensure(project Project) error {
	r.logger.Log("debug", "ensuring chart is installed", "name", project.Name, "ref", project.Ref)

	var err error

	// We create a tmp dir in which all Helm values files and tarballs are written
	// to. After we are done we can just remove the whole tmp dir to clean up.
	var tmpDir string
	{
		tmpDir, err = afero.TempDir(r.fileSystem, "", "draughtsman-operator-project-resource")
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			err := r.fileSystem.RemoveAll(tmpDir)
			if err != nil {
				r.logger.Log("error", fmt.Sprintf("could not remove tmp dir: %#v", err), "name", project.Name, "ref", project.Ref)
			}
		}()
	}

	var tarballPath string
	{
		tarballPath = path.Join(tmpDir, r.tarballName(project))

		_, err := r.runHelmCommand("pull", "registry", "pull", "--dest", tmpDir, "--tarball", r.versionedChartName(project))
		if err != nil {
			return microerror.Mask(err)
		}
		exists, err := afero.Exists(r.fileSystem, tarballPath)
		if !exists {
			return microerror.Maskf(helmError, "could not find downloaded tarball at '%s'", tarballPath)
		}

		r.logger.Log("debug", "downloaded chart", "tarball", tarballPath)
	}

	// The intaller accepts multiple configurers during initialization. Here we
	// iterate over all of them to get all the values they provide. For each
	// values file we have to create a file in the tmp dir we created above.
	var valuesFilesArgs []string
	for _, c := range r.configurers {
		fileName := filepath.Join(tmpDir, fmt.Sprintf("%s-values.yaml", strings.ToLower(string(c.Type()))))
		values, err := c.Values()
		if err != nil {
			return microerror.Mask(err)
		}

		err = afero.WriteFile(r.fileSystem, fileName, []byte(values), os.FileMode(0644))
		if err != nil {
			return microerror.Mask(err)
		}

		valuesFilesArgs = append(valuesFilesArgs, "--values", fileName)
	}

	// The arguments used to execute Helm for app installation can take multiple
	// values files. At the end the command looks something like this.
	//
	//     helm upgrade --install --values ${file1} --values $(file2) ${project} ${tarball_path}
	//
	var installCommand []string
	{
		installCommand = append(installCommand, "upgrade", "--install")
		installCommand = append(installCommand, valuesFilesArgs...)
		installCommand = append(installCommand, project.Name, tarballPath)

		_, err := r.runHelmCommand("install", installCommand...)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

// login logs the configured user into the configured registry.
func (r *Resource) login() error {
	r.logger.Log("debug", "logging into registry", "username", r.username, "registry", r.registry)

	_, err := r.runHelmCommand(
		"login",
		"registry",
		"login",
		fmt.Sprintf("--user=%v", r.username),
		fmt.Sprintf("--password=%v", r.password),
		r.registry,
	)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// runHelmCommand runs the given Helm command.
func (r *Resource) runHelmCommand(name string, args ...string) ([]byte, error) {
	r.logger.Log("debug", "running helm command", "name", name)

	defer updateHelmMetrics(name, time.Now())

	cmd := exec.Command(r.helmBinaryPath, args...)
	b, err := cmd.CombinedOutput()
	r.logger.Log("debug", "ran helm command", "command", name, "output", string(b))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if strings.Contains(string(b), "Error") {
		return nil, microerror.Maskf(helmError, string(b))
	}

	return b, nil
}

// tarballName builds a tarball name, given a project name and sha.
func (r *Resource) tarballName(project Project) string {
	return fmt.Sprintf(
		tarballNameFormat,
		r.organisation,
		project.Name,
		project.Ref,
	)
}

// versionedChartName builds a chart name, including a version,
// given a project name and a sha.
func (r *Resource) versionedChartName(project Project) string {
	return fmt.Sprintf(
		versionedChartFormat,
		r.registry,
		r.organisation,
		project.Name,
		project.Ref,
	)
}

func containsProject(list []Project, item Project) bool {
	for _, l := range list {
		if l.Name == item.Name {
			return true
		}
	}

	return false
}

func toCustomObject(v interface{}) (draughtsmantpr.CustomObject, error) {
	customObjectPointer, ok := v.(*draughtsmantpr.CustomObject)
	if !ok {
		return draughtsmantpr.CustomObject{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &draughtsmantpr.CustomObject{}, v)
	}
	customObject := *customObjectPointer

	return customObject, nil
}

func toProjects(v interface{}) ([]Project, error) {
	projects, ok := v.([]Project)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", []Project{}, v)
	}

	return projects, nil
}
