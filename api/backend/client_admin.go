package backend

// AdminUser is the domain representation of a Bitbucket Server / DC user as
// returned by the admin users endpoint.
type AdminUser struct {
	Slug        string
	DisplayName string
	Email       string
	Active      bool
	Type        string // "NORMAL", "SERVICE", etc.
}

// AdminLicense holds the license details for a Bitbucket Server / DC instance.
type AdminLicense struct {
	Tier          string
	Users         int
	ServerId      string
	License       string
	ExpiryDate    string
	SupportExpiry string
	CreationDate  string
}

// ClusterNode represents a node in a Bitbucket Server / DC cluster.
type ClusterNode struct {
	NodeId  string
	Name    string
	Address string
	State   string
	Local   bool
}

// MailServerConfig holds the mail-server settings for a Bitbucket Server / DC instance.
type MailServerConfig struct {
	Hostname        string
	Port            int
	Protocol        string // "smtp" | "smtps"
	UseStartTLS     bool
	RequireStartTLS bool
	Username        string
	SenderAddress   string
	// Password is write-only — not returned by GET; include for the set command.
	// The json:"-" tag ensures it is never serialised in --json output.
	Password string `json:"-"`
}

// BannerConfig holds the site-wide announcement banner settings for a
// Bitbucket Server / DC instance.
type BannerConfig struct {
	Message  string `json:"message"`
	Audience string `json:"audience"` // "ALL" | "AUTHENTICATED" | "UNAUTHENTICATED"
	Enabled  bool   `json:"enabled"`
}

// AdminClient exposes Bitbucket Server / Data Center administration
// operations. Bitbucket Cloud does not expose these endpoints — calls against
// Cloud return ErrUnsupportedOnHost via AsAdminClient.
type AdminClient interface {
	RotateSecrets() error
	GetLoggingConfig() (LoggingConfig, error)
	SetLoggingConfig(in LoggingConfigInput) error

	// User management (Server/DC only)
	ListAdminUsers(filter string, limit int) ([]AdminUser, error)
	RenameUser(slug, newSlug string) error
	ActivateUser(slug string) error
	DeactivateUser(slug string) error

	// System info (Server/DC only)
	GetLicense() (AdminLicense, error)
	GetClusterNodes() ([]ClusterNode, error)

	// Mail server config (Server/DC only)
	GetMailServerConfig() (MailServerConfig, error)
	SetMailServerConfig(in MailServerConfig) error

	// Banner config (Server/DC only)
	GetBanner() (BannerConfig, error)
	SetBanner(in BannerConfig) error
	ClearBanner() error

	// Rate-limit config (Server/DC only)
	GetRateLimitConfig() (RateLimitConfig, error)
	SetRateLimitConfig(in RateLimitConfig) error
}

// FeatureAdmin names the admin capability.
const FeatureAdmin Feature = "admin"

// AsAdminClient returns the AdminClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a backend that
// doesn't implement admin operations (currently Bitbucket Cloud).
func AsAdminClient(c Client, host string) (AdminClient, error) {
	return requireFeature[AdminClient](c, host, specFor(FeatureAdmin))
}
