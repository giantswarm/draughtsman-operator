package spec

// NotifierType represents the type of Notifier to configure.
type NotifierType string

type Project struct {
	Name string
	Ref  string
}

// Notifier represents a Service that notifies of install status.
type Notifier interface {
	// Failed notifies of failed installations.
	Failed(project Project, errorMessage string) error
	// Success notifies of successful installations.
	Success(project Project) error
}
