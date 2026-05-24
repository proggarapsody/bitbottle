package backend

// RepoPRSettingsClient reads and writes the per-repo pull-request gate
// configuration. This capability is Server-only; Cloud returns
// ErrUnsupportedOnHost for both methods.
type RepoPRSettingsClient interface {
	GetRepoPRSettings(ns, slug string) (RepoPRSettings, error)
	UpdateRepoPRSettings(ns, slug string, in RepoPRSettingsInput) (RepoPRSettings, error)
}

// RepoPRSettings is the domain representation of per-repo PR gate settings.
type RepoPRSettings struct {
	RequiredApprovers        int      `json:"requiredApprovers"`
	RequiredAllApprovers     bool     `json:"requiredAllApprovers"`
	RequiredAllTasksComplete bool     `json:"requiredAllTasksComplete"`
	RequiredSuccessfulBuilds int      `json:"requiredSuccessfulBuilds"`
	MergeStrategy            string   `json:"mergeStrategy"`
	AllowedStrategies        []string `json:"allowedStrategies"`
}

// RepoPRSettingsInput carries the fields to update. Nil pointer = leave unchanged.
type RepoPRSettingsInput struct {
	RequiredApprovers        *int
	RequiredAllApprovers     *bool
	RequiredAllTasksComplete *bool
	RequiredSuccessfulBuilds *int
	MergeStrategy            *string
	AllowedStrategies        *[]string
}

// FeatureRepoPRSettings names the repo-pr-settings capability for typed-error reporting.
const FeatureRepoPRSettings Feature = "repo_pr_settings"

// AsRepoPRSettingsClient returns the RepoPRSettingsClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the RepoPRSettings capability.
func AsRepoPRSettingsClient(c Client, host string) (RepoPRSettingsClient, error) {
	return requireFeature[RepoPRSettingsClient](c, host, specFor(FeatureRepoPRSettings))
}
