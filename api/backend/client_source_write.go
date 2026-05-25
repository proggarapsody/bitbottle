package backend

// SourceWriter can create or update a file on a branch via the API.
type SourceWriter interface {
	PutFile(ns, slug, path string, in PutFileInput) error
}

// PutFileInput holds the parameters for creating or updating a file.
type PutFileInput struct {
	Content      string // raw file content
	Branch       string // target branch name
	Message      string // commit message
	SourceCommit string // optional: expected HEAD SHA for conflict detection
}

// FeatureSourceWrite names the source-write capability.
const FeatureSourceWrite Feature = "SourceWriter"

// AsSourceWriter returns the SourceWriter view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a backend
// that doesn't implement source-write operations.
func AsSourceWriter(c Client, host string) (SourceWriter, error) {
	return requireFeature[SourceWriter](c, host, specFor(FeatureSourceWrite))
}
