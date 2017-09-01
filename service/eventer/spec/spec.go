package spec

// EventerType represents the type of Eventer to configure.
type EventerType string

// Eventer represents a Service that checks for deployment events.
type Eventer interface {
	// SetFailedStatus updates the given DeploymentEvent's remote state to a
	// failed state.
	SetFailedStatus(event DeploymentEvent) error
	// SetSuccessStatus updates the given DeploymentEvent's remote state to a
	// success state.
	SetSuccessStatus(event DeploymentEvent) error
}

// DeploymentEvent represents a request for a chart to be deployed.
type DeploymentEvent struct {
	// ID is an identifier for the deployment event.
	ID int
	// Name is the name of the project of the chart to deploy, e.g: aws-operator.
	Name string
	// Sha is the version of the chart to deploy.
	Sha string
}
