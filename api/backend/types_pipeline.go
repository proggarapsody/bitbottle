package backend

// Pipeline is the domain representation of a Bitbucket Cloud pipeline run.
type Pipeline struct {
	UUID        string
	BuildNumber int
	State       string // PENDING, IN_PROGRESS, SUCCESSFUL, FAILED, ERROR, STOPPED
	RefType     string // "branch", "tag", "commit"
	RefName     string
	CreatedOn   string
	Duration    int // seconds
	WebURL      string
}

// RunPipelineInput carries the parameters for triggering a pipeline run.
type RunPipelineInput struct {
	Branch string
}

// PipelineStep is the domain representation of a single step within a
// Bitbucket Cloud pipeline run.
type PipelineStep struct {
	UUID     string
	Name     string
	State    string // PENDING, IN_PROGRESS, SUCCESSFUL, FAILED, ERROR, STOPPED
	Result   string // populated when State has flattened from COMPLETED
	Duration int    // seconds
}

// PipelineVariable is a repository-level pipeline variable on Bitbucket Cloud.
// Value is empty when Secured is true (the API never returns secured values).
type PipelineVariable struct {
	UUID    string
	Key     string
	Value   string
	Secured bool
}

// PipelineVariableInput carries the parameters for upserting a pipeline
// variable by Key.
type PipelineVariableInput struct {
	Key     string
	Value   string
	Secured bool
}

// PipelineTriggerInput carries the parameters for triggering a pipeline via
// the PipelineTriggerClient interface. Variables supplements the per-run
// environment; each entry maps to a Bitbucket pipeline variable object.
type PipelineTriggerInput struct {
	Branch    string
	Variables []PipelineVariable
}

// PipelineTriggerResult is returned by TriggerPipeline on success.
type PipelineTriggerResult struct {
	UUID  string `json:"uuid"`
	State string `json:"state"`
	Link  string `json:"link"`
}

// PipelineSchedule is the domain representation of a Bitbucket Cloud pipeline
// schedule.
type PipelineSchedule struct {
	UUID           string `json:"uuid"`
	Enabled        bool   `json:"enabled"`
	CronExpression string `json:"cronExpression"`
	Branch         string `json:"branch"`
}

// PipelineScheduleInput carries the parameters for creating a pipeline
// schedule.
type PipelineScheduleInput struct {
	CronExpression string
	Branch         string
	Enabled        bool
}

// PipelineCache is a named cache entry for a Bitbucket Cloud pipeline.
type PipelineCache struct {
	UUID          string `json:"uuid"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	FileSizeBytes int64  `json:"fileSizeBytes"`
	CreatedOn     string `json:"createdOn"`
}

// PipelineOIDCConfig is the OIDC discovery document published by Bitbucket
// Cloud Pipelines for workload-identity federation.
type PipelineOIDCConfig struct {
	Issuer                 string   `json:"issuer"`
	JWKSURI                string   `json:"jwks_uri"`
	SubjectTypesSupported  []string `json:"subject_types_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
	ClaimsSupported        []string `json:"claims_supported"`
}

// PipelineOIDCKey is a single JSON Web Key from the Bitbucket Cloud Pipelines
// JWKS endpoint.
type PipelineOIDCKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// PipelineOIDCKeys is the JWKS key-set published by Bitbucket Cloud Pipelines.
type PipelineOIDCKeys struct {
	Keys []PipelineOIDCKey `json:"keys"`
}
