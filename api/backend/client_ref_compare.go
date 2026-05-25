package backend

// RefComparison is the domain representation of the result of comparing two
// refs (branches or commits). AheadBy is the number of commits that Head has
// ahead of Base; BehindBy is the number of commits that Base has ahead of Head.
type RefComparison struct {
	Base          string   `json:"base"`
	Head          string   `json:"head"`
	AheadBy       int      `json:"ahead_by"`
	BehindBy      int      `json:"behind_by"`
	CommitsAhead  []Commit `json:"commits_ahead"`
	CommitsBehind []Commit `json:"commits_behind"`
}

// RefComparer is implemented by Cloud and Server/DC backends that support
// branch comparison.
type RefComparer interface {
	CompareRefs(ns, slug, base, head string, limit int) (RefComparison, error)
}

// FeatureRefCompare names the ref-compare capability.
const FeatureRefCompare Feature = "ref_compare"

// AsRefComparer returns the RefComparer view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) if the backend does not implement the capability.
func AsRefComparer(c Client, host string) (RefComparer, error) {
	return requireFeature[RefComparer](c, host, specFor(FeatureRefCompare))
}
