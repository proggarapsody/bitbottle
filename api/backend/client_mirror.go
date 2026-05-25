package backend

// MirrorClient is implemented only by Bitbucket Server/DC clients. Bitbucket
// Cloud has no Smart Mirror API, so the entire mirror surface is gated behind
// AsMirrorClient.
type MirrorClient interface {
	ListMirrorServers(limit int) ([]MirrorServer, error)
	GetMirrorServer(id string) (MirrorServer, error)
	ListMirroredRepos(mirrorID string, limit int) ([]MirroredRepo, error)
}

// FeatureMirror names the mirror capability for typed-error reporting.
const FeatureMirror Feature = "mirror"

// AsMirrorClient returns the MirrorClient view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Cloud backend.
func AsMirrorClient(c Client, host string) (MirrorClient, error) {
	return requireFeature[MirrorClient](c, host, specFor(FeatureMirror))
}
