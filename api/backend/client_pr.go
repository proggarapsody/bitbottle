package backend

import "fmt"

// PRLister lists pull requests.
type PRLister interface {
	ListPRs(ns, slug, state string, limit int) ([]PullRequest, error)
}

// PRReader reads a single pull request.
type PRReader interface {
	GetPR(ns, slug string, id int) (PullRequest, error)
}

// PRCreator creates a pull request.
type PRCreator interface {
	CreatePR(ns, slug string, in CreatePRInput) (PullRequest, error)
}

// PRMerger merges a pull request.
type PRMerger interface {
	MergePR(ns, slug string, id int, in MergePRInput) (PullRequest, error)
}

// PRApprover approves a pull request.
type PRApprover interface {
	ApprovePR(ns, slug string, id int) error
}

// PRDiffer returns the unified diff for a pull request.
type PRDiffer interface {
	GetPRDiff(ns, slug string, id int) (string, error)
}

// PREditor updates a pull request's title or description.
type PREditor interface {
	UpdatePR(ns, slug string, id int, in UpdatePRInput) (PullRequest, error)
}

// PRDecliner declines a pull request.
type PRDecliner interface {
	DeclinePR(ns, slug string, id int) error
}

// PRReopener reverses a decline. Implemented only by Bitbucket Server / Data
// Center (POST /pull-requests/{id}/reopen). Bitbucket Cloud has no reopen
// primitive — declined PRs are terminal there (BCLOUD-23807) — so callers
// route through AsPRReopener so the Cloud-only constraint surfaces as a
// typed ErrUnsupportedOnHost rather than a panic.
type PRReopener interface {
	ReopenPR(ns, slug string, id int) error
}

// PRUnapprover removes approval from a pull request.
type PRUnapprover interface {
	UnapprovePR(ns, slug string, id int) error
}

// PRReadier marks a draft pull request as ready for review.
type PRReadier interface {
	ReadyPR(ns, slug string, id int) error
}

// PRReviewRequester requests reviewers on a pull request.
type PRReviewRequester interface {
	RequestReview(ns, slug string, id int, users []string) error
}

// PRChangesRequester can request changes on a pull request (Cloud only).
// Access via type assertion — not embedded in Client.
type PRChangesRequester interface {
	RequestChangesPR(ns, slug string, id int) error
}

// PRCommentLister lists top-level comments on a pull request.
type PRCommentLister interface {
	ListPRComments(ns, slug string, id int) ([]PRComment, error)
}

// PRCommentAdder adds a general comment to a pull request.
type PRCommentAdder interface {
	AddPRComment(ns, slug string, id int, in AddPRCommentInput) (PRComment, error)
}

// FeaturePRReopen names the PR-reopen capability for typed-error reporting.
// Bitbucket Cloud has no reopen primitive (BCLOUD-23807), so callers gate
// the feature behind AsPRReopener.
const FeaturePRReopen Feature = "pr-reopen"

// AsPRReopener returns the PRReopener view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when the backend at host has no reopen
// primitive (currently Bitbucket Cloud).
func AsPRReopener(c Client, host string) (PRReopener, error) {
	r, ok := c.(PRReopener)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePRReopen),
			Message: fmt.Sprintf("pr reopen is not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return r, nil
}
