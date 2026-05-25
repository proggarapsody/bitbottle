package backend

// RunnerClient manages Bitbucket Cloud Pipelines self-hosted runners.
type RunnerClient interface {
	ListRunners(workspace string) ([]Runner, error)
	CreateRunner(workspace string, in CreateRunnerInput) (Runner, error)
	DeleteRunner(workspace, runnerUUID string) error
}

// Runner represents a Bitbucket Cloud Pipelines self-hosted runner.
type Runner struct {
	UUID      string
	Name      string
	State     string // "ONLINE" | "OFFLINE" | "DISABLED"
	Platform  RunnerPlatform
	Labels    []string
	CreatedOn string
}

// RunnerPlatform describes the OS and architecture of a runner.
type RunnerPlatform struct {
	Operating string // "LINUX" | "WINDOWS" | "MACOS"
	Arch      string // "AMD64" | "ARM64"
}

// CreateRunnerInput is the input for creating a new runner.
type CreateRunnerInput struct {
	Name     string
	Labels   []string
	Platform RunnerPlatform
}

// FeatureRunner is the feature key for RunnerClient.
const FeatureRunner Feature = "Runner"

// AsRunnerClient asserts RunnerClient capability on c. Returns a typed
// DomainError{ErrUnsupportedOnHost} when the backend does not support it.
func AsRunnerClient(c Client, host string) (RunnerClient, error) {
	return requireFeature[RunnerClient](c, host, specFor(FeatureRunner))
}
