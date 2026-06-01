package backend

import "encoding/json"

// RepoHookClient manages per-repository plugin hook script settings
// on Bitbucket Server / Data Center.
type RepoHookClient interface {
	ListRepoHooks(project, slug string) ([]RepoHook, error)
	GetRepoHook(project, slug, hookKey string) (RepoHook, error)
	EnableRepoHook(project, slug, hookKey string) error
	DisableRepoHook(project, slug, hookKey string) error
	GetRepoHookSettings(project, slug, hookKey string) (json.RawMessage, error)
	SetRepoHookSettings(project, slug, hookKey string, cfg json.RawMessage) error
}

// RepoHook represents a plugin hook script installed on a repository.
type RepoHook struct {
	Key        string
	Name       string
	Version    string
	Enabled    bool
	Configured bool
}

// FeatureRepoHooks names the repo hook scripts capability for typed-error
// reporting.
const FeatureRepoHooks Feature = "repo-hooks"

// AsRepoHookClient returns the RepoHookClient view of c, or a typed
// ErrUnsupportedOnHost when c does not implement it (e.g. Cloud).
func AsRepoHookClient(c Client, host string) (RepoHookClient, error) {
	return requireFeature[RepoHookClient](c, host, specFor(FeatureRepoHooks))
}
