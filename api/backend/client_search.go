package backend

// CodeSearcher performs workspace-scoped code search on Bitbucket Cloud.
// Bitbucket Server / Data Center does not expose a first-class REST code-
// search endpoint (search there is provided by the separate Sourcegraph
// integration or third-party plugins), so the entire surface is gated
// behind AsCodeSearcher. Server adapters intentionally do NOT implement
// this interface — the type-assertion in AsCodeSearcher is what surfaces
// the typed host.unsupported error to callers.
type CodeSearcher interface {
	SearchCode(workspace, query string, limit int) ([]CodeSearchHit, error)
}

// FeatureCodeSearch names the code-search capability for typed-error
// reporting.
const FeatureCodeSearch Feature = "code-search"

// AsCodeSearcher returns the CodeSearcher view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a backend that doesn't
// model code search (currently Server/DC).
func AsCodeSearcher(c Client, host string) (CodeSearcher, error) {
	return requireFeature[CodeSearcher](c, host, specFor(FeatureCodeSearch))
}
