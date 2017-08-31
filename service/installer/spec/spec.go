package spec

// InstallerType represents the type of Installer to configure.
type InstallerType string

type Project struct {
	Name string
	Ref  string
}

// Installer represents a Service that installs charts.
type Installer interface {
	// Install takes a Project, and installs the referenced chart. If an error
	// occurs, the returned error will be non-nil.
	Install(project Project) error
	// List returns a list of installed charts in form of a list of Project items.
	List() ([]Project, error)
}