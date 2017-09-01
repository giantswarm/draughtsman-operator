package spec

type Project struct {
	// ID is the deployment event ID the current project is associated with.
	ID string `json:"id" yaml:"id"`
	// Name is the project name being deployed.
	Name string `json:"name" yaml:"name"`
	// Ref is the ref/sha acting as version. This is usually a Git commit hash.
	Ref string `json:"ref" yaml:"ref"`
}
