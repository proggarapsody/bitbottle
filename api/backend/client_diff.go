package backend

// DiffClient is implemented by both Cloud and Server backends.
type DiffClient interface {
	GetDiff(ns, slug, from, to string) (string, error)
	GetDiffStat(ns, slug, from, to string) (DiffStat, error)
}

// FeatureDiff names the diff capability for typed-error reporting.
const FeatureDiff Feature = "diff"

// AsDiffClient returns the DiffClient view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) if the backend at host does not implement the
// Diff capability.
func AsDiffClient(c Client, host string) (DiffClient, error) {
	return requireFeature[DiffClient](c, host, specFor(FeatureDiff))
}
