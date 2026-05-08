package backend

import "fmt"

// DefaultReviewersResolver looks up the configured "default reviewers" for a
// repository given a source/target ref pair. Only Bitbucket Server / Data
// Center exposes this — Cloud has a similar feature with a different,
// non-trivial wire shape that we don't model yet.
//
// Implementations may need additional context (e.g. numeric repo ID) which
// they obtain themselves; callers pass only the values they natively know.
type DefaultReviewersResolver interface {
	DefaultReviewers(ns, slug, fromBranch, toBranch string) ([]User, error)
}

// FeatureDefaultReviewers names the default-reviewers capability for
// typed-error reporting.
const FeatureDefaultReviewers Feature = "default-reviewers"

// AsDefaultReviewersResolver returns the DefaultReviewersResolver view of c,
// or a typed *DomainError when the backend doesn't implement it (currently
// Cloud). Callers use the returned error to decide whether to skip the
// auto-apply step entirely.
func AsDefaultReviewersResolver(c Client, host string) (DefaultReviewersResolver, error) {
	r, ok := c.(DefaultReviewersResolver)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureDefaultReviewers),
			Message: fmt.Sprintf("default reviewers lookup is not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return r, nil
}
