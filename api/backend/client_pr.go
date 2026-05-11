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

// PRReviewer submits a compound review on a pull request — an optional
// top-level body comment, zero or more inline comments, and a review action
// (approve / request_changes / comment). The adapter sequences the writes
// internally so callers express "what the review looks like" rather than
// "how to assemble it".
type PRReviewer interface {
	SubmitReview(ns, slug string, id int, in SubmitReviewInput) error
}

// PRCommentLister lists top-level comments on a pull request.
type PRCommentLister interface {
	ListPRComments(ns, slug string, id int) ([]PRComment, error)
}

// PRCommentAdder adds a general comment to a pull request.
type PRCommentAdder interface {
	AddPRComment(ns, slug string, id int, in AddPRCommentInput) (PRComment, error)
}

// PRCommentEditor updates the body of an existing comment on a pull request.
type PRCommentEditor interface {
	EditPRComment(ns, slug string, id, commentID int, body string) (PRComment, error)
}

// PRCommentDeleter removes a comment from a pull request.
type PRCommentDeleter interface {
	DeletePRComment(ns, slug string, id, commentID int) error
}

// PRCommentResolver marks a PR comment as resolved. Implemented only by
// Bitbucket Cloud (whose comments carry a native `resolution` field).
// Bitbucket Server has no equivalent concept on regular comments — the
// closest analogue lives on tasks, which is a separate feature scope —
// so callers route through AsPRCommentResolver and surface the constraint
// as a typed ErrUnsupportedOnHost.
type PRCommentResolver interface {
	ResolvePRComment(ns, slug string, id, commentID int) error
}

// PRActivityReader reads the activity event stream for a pull request.
type PRActivityReader interface {
	GetPRActivity(ns, slug string, id int, limit int) ([]PRActivityEvent, error)
}

// FeaturePRCommentResolve names the inline-comment resolution capability
// for typed-error reporting via AsPRCommentResolver.
const FeaturePRCommentResolve Feature = "pr-comment-resolve"

// AsPRCommentResolver returns the PRCommentResolver view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when the backend at host has no
// resolution primitive (currently Bitbucket Server / Data Center).
func AsPRCommentResolver(c Client, host string) (PRCommentResolver, error) {
	r, ok := c.(PRCommentResolver)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePRCommentResolve),
			Message: fmt.Sprintf("pr comment resolve is not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return r, nil
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
