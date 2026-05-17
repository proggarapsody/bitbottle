package backend

// Deployment is the domain representation of a Bitbucket Cloud deployment.
// State values: PENDING, IN_PROGRESS, COMPLETED, STOPPED, FAILED.
type Deployment struct {
	UUID        string
	State       string
	Environment Environment
	Release     struct {
		Name       string
		URL        string
		CommitHash string
	}
}

// Environment is a Bitbucket Cloud deployment environment (e.g. Test, Staging, Production).
// Rank is the numeric ordering position.
type Environment struct {
	UUID string
	Name string
	Type string // Test | Staging | Production
	Rank int
}

// CreateEnvironmentInput carries the parameters for creating a deployment environment.
type CreateEnvironmentInput struct {
	Name string
	Type string
	Rank int
}

// EnvVariable is a deployment-environment variable on Bitbucket Cloud.
// Value is empty when Secured is true (the API never returns secured values).
type EnvVariable struct {
	UUID    string
	Key     string
	Value   string // empty if Secured
	Secured bool
}

// EnvVariableInput carries the parameters for creating or updating an
// environment variable.
type EnvVariableInput struct {
	Key     string
	Value   string
	Secured bool
}
