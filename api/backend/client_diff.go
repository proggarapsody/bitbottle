package backend

import "fmt"

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
	dc, ok := c.(DiffClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureDiff),
			Message: fmt.Sprintf("diff is not supported on %s", host),
		}
	}
	return dc, nil
}
