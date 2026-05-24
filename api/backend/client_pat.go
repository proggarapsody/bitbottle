package backend

import "time"

// PATClient is implemented by Bitbucket Server/DC only. Bitbucket Cloud API
// token management is not available via the REST API.
type PATClient interface {
	ListPATs(userSlug string, limit int) ([]PAT, error)
	CreatePAT(userSlug string, in CreatePATInput) (PATWithSecret, error)
	RevokePAT(userSlug, tokenID string) error
}

// PAT is a Bitbucket Server/DC personal access token (read-only fields).
type PAT struct {
	ID          string
	Name        string
	Permissions []string
	CreatedDate time.Time
	ExpiryDate  *time.Time
	LastUsed    *time.Time
}

// PATWithSecret is returned only by CreatePAT.
// Token is the raw secret — display once, never log.
type PATWithSecret struct {
	PAT
	Token string
}

// CreatePATInput carries the parameters for creating a PAT.
type CreatePATInput struct {
	Name        string
	Permissions []string // REPO_READ, REPO_WRITE, PR_READ, PR_WRITE, PROJECT_READ, PROJECT_WRITE
	ExpiryDays  *int     // nil = no expiry
}

// FeaturePAT names the PAT management capability.
const FeaturePAT Feature = "pat"

// AsPATClient returns the PATClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// support PAT management.
func AsPATClient(c Client, host string) (PATClient, error) {
	return requireFeature[PATClient](c, host, specFor(FeaturePAT))
}
