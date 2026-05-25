package backend

// IPAllowlistClient manages Bitbucket Cloud workspace IP allowlists.
type IPAllowlistClient interface {
	ListIPAllowlists(workspace string) ([]IPAllowlist, error)
	CreateIPAllowlist(workspace string, in CreateIPAllowlistInput) (IPAllowlist, error)
	DeleteIPAllowlist(workspace, entryUUID string) error
}

// IPAllowlist is a single IP allowlist entry for a workspace.
type IPAllowlist struct {
	UUID        string `json:"uuid"`
	CIDR        string `json:"cidr"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// CreateIPAllowlistInput carries parameters for adding an allowlist entry.
type CreateIPAllowlistInput struct {
	CIDR        string
	Description string
	Enabled     bool
}

// FeatureIPAllowlist is the feature key for IPAllowlistClient.
const FeatureIPAllowlist Feature = "IPAllowlist"

// AsIPAllowlistClient asserts IPAllowlistClient capability. Returns a typed
// DomainError{ErrUnsupportedOnHost} when the backend does not support it.
func AsIPAllowlistClient(c Client, host string) (IPAllowlistClient, error) {
	return requireFeature[IPAllowlistClient](c, host, specFor(FeatureIPAllowlist))
}
