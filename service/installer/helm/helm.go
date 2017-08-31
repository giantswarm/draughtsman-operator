package helm

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"

	configurerspec "github.com/giantswarm/draughtsman-operator/service/configurer/spec"
	"github.com/giantswarm/draughtsman-operator/service/installer/spec"
)

const (
	// tarballNameFormat is the format for the name of the chart tarball.
	tarballNameFormat = "%v_%v-chart_1.0.0-%v.tar.gz"
	// versionedChartFormat is the format the CNR registry uses to address
	// charts. For example, we use this to address that chart to pull.
	versionedChartFormat = "%v/%v/%v-chart@1.0.0-%v"
)

// HelmInstallerType is an Installer that uses Helm.
var HelmInstallerType spec.InstallerType = "HelmInstaller"

// Config represents the configuration used to create a Helm Installer.
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

// DefaultConfig provides a default configuration to create a new Helm Installer
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Configurers: nil,
		FileSystem:  nil,
		Logger:      nil,

		// Settings.
		HelmBinaryPath: "",
		Organisation:   "",
		Password:       "",
		Registry:       "",
		Username:       "",
	}
}

// Installer is an implementation of the Installer interface, that uses Helm to
// install charts.
type Installer struct {
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

// New creates a new configured Helm Installer.
func New(config Config) (*Installer, error) {
	// Dependencies.
	if config.Configurers == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Configurers must not be empty")
	}
	if config.FileSystem == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.FileSystem must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	if config.HelmBinaryPath == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.HelmBinaryPath must not be empty")
	}
	if config.Organisation == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Organisation must not be empty")
	}
	if config.Password == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Password must not be empty")
	}
	if config.Registry == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Registry must not be empty")
	}
	if config.Username == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.Username must not be empty")
	}

	exists, err := afero.Exists(config.FileSystem, config.HelmBinaryPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if !exists {
		return nil, microerror.Maskf(invalidConfigError, "helm binary path '%s' does not exist", config.HelmBinaryPath)
	}

	installer := &Installer{
		// Dependencies.
		configurers: config.Configurers,
		fileSystem:  config.FileSystem,
		logger:      config.Logger,

		// Settings.
		helmBinaryPath: config.HelmBinaryPath,
		organisation:   config.Organisation,
		password:       config.Password,
		registry:       config.Registry,
		username:       config.Username,
	}

	if err := installer.login(); err != nil {
		return nil, microerror.Mask(err)
	}

	return installer, nil
}

func (i *Installer) Install(project spec.Project) error {
	i.logger.Log("debug", "ensuring chart is installed", "name", project.Name, "ref", project.Ref)

	var err error

	// We create a tmp dir in which all Helm values files and tarballs are written
	// to. After we are done we can just remove the whole tmp dir to clean up.
	var tmpDir string
	{
		tmpDir, err = afero.TempDir(i.fileSystem, "", "draughtsman-operator-helm-installer")
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			err := i.fileSystem.RemoveAll(tmpDir)
			if err != nil {
				i.logger.Log("error", fmt.Sprintf("could not remove tmp dir: %#v", err), "name", project.Name, "ref", project.Ref)
			}
		}()
	}

	var tarballPath string
	{
		tarballPath = path.Join(tmpDir, i.tarballName(project))

		_, err := i.runHelmCommand("pull", "registry", "pull", "--dest", tmpDir, "--tarball", i.versionedChartName(project))
		if err != nil {
			return microerror.Mask(err)
		}
		exists, err := afero.Exists(i.fileSystem, tarballPath)
		if !exists {
			return microerror.Maskf(helmError, "could not find downloaded tarball at '%s'", tarballPath)
		}

		i.logger.Log("debug", "downloaded chart", "tarball", tarballPath)
	}

	// The intaller accepts multiple configurers during initialization. Here we
	// iterate over all of them to get all the values they provide. For each
	// values file we have to create a file in the tmp dir we created above.
	var valuesFilesArgs []string
	for _, c := range i.configurers {
		fileName := filepath.Join(tmpDir, fmt.Sprintf("%s-values.yaml", strings.ToLower(string(c.Type()))))
		values, err := c.Values()
		if err != nil {
			return microerror.Mask(err)
		}

		err = afero.WriteFile(i.fileSystem, fileName, []byte(values), os.FileMode(0644))
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

		_, err := i.runHelmCommand("install", installCommand...)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (i *Installer) List(projects []spec.Project) ([]spec.Project, error) {
	b, err := i.runHelmCommand("list", "list")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	listProjects, err := bytesToProjects(b)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var filteredProjects []spec.Project

	for _, p := range listProjects {
		foundProject, err := getProjectFromList(projects, p)
		if IsNotFound(err) {
			continue
		}

		filteredProjects = append(filteredProjects, foundProject)
	}

	return filteredProjects, nil
}

// login logs the configured user into the configured registry.
func (i *Installer) login() error {
	i.logger.Log("debug", "logging into registry", "username", i.username, "registry", i.registry)

	_, err := i.runHelmCommand(
		"login",
		"registry",
		"login",
		fmt.Sprintf("--user=%v", i.username),
		fmt.Sprintf("--password=%v", i.password),
		i.registry,
	)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// runHelmCommand runs the given Helm command.
func (i *Installer) runHelmCommand(name string, args ...string) ([]byte, error) {
	i.logger.Log("debug", "running helm command", "name", name)

	defer updateHelmMetrics(name, time.Now())

	cmd := exec.Command(i.helmBinaryPath, args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	i.logger.Log("debug", "ran helm command", "command", name, "output", string(b))

	if strings.Contains(string(b), "Error") {
		return nil, microerror.Maskf(helmError, string(b))
	}

	return b, nil
}

// tarballName builds a tarball name, given a project name and sha.
func (i *Installer) tarballName(project spec.Project) string {
	return fmt.Sprintf(
		tarballNameFormat,
		i.organisation,
		project.Name,
		project.Ref,
	)
}

// versionedChartName builds a chart name, including a version,
// given a project name and a sha.
func (i *Installer) versionedChartName(project spec.Project) string {
	return fmt.Sprintf(
		versionedChartFormat,
		i.registry,
		i.organisation,
		project.Name,
		project.Ref,
	)
}

// bytesToProjects parses projects from the given bytes.
//
// NOTE that the retruned list of projects does eventually contain incomplete
// ref/sha information. This is because of certain helm limitations when listing
// charts.
func bytesToProjects(b []byte) ([]spec.Project, error) {
	var list []spec.Project

	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t == "" {
			continue
		}
		f := strings.Fields(t)

		if len(f) == 0 {
			continue
		}
		if f[0] == "NAME" && f[1] == "REVISION" {
			// Here we assume we have the list header. we are not interested in it. So
			// we ignore it.
			continue
		}

		split := strings.Split(f[8], "-")
		if len(split) == 0 {
			continue
		}
		regEx := regexp.MustCompile("^[0-9a-z]{3,}")
		ref := regEx.FindString(split[len(split)-1])

		list = append(list, spec.Project{Name: f[0], Ref: ref})
	}
	err := scanner.Err()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return list, nil
}

func getProjectFromList(list []spec.Project, p spec.Project) (spec.Project, error) {
	for _, l := range list {
		if l.Name == p.Name && strings.HasPrefix(l.Ref, p.Ref) {
			return l, nil
		}
	}

	return spec.Project{}, microerror.Maskf(notFoundError, p.Name)
}
