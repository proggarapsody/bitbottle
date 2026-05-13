package backend

import "fmt"

// CommitCommentReactor manages emoji reactions on commit comments.
// Implemented only by Bitbucket Server / Data Center. Bitbucket Cloud has no
// equivalent API. Callers route through AsCommitCommentReactor to surface the
// constraint as a typed ErrUnsupportedOnHost.
type CommitCommentReactor interface {
	ListCommitCommentReactions(ns, slug, hash string, commentID int) ([]CommentReaction, error)
	AddCommitCommentReaction(ns, slug, hash string, commentID int, emoji string) error
	RemoveCommitCommentReaction(ns, slug, hash string, commentID int, emoji string) error
}

// FeatureCommitCommentReactions names the commit comment reaction capability.
const FeatureCommitCommentReactions Feature = "commit-comment-reactions"

// AsCommitCommentReactor returns the CommitCommentReactor view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when the backend at host has no
// reaction primitive (currently Bitbucket Cloud).
func AsCommitCommentReactor(c Client, host string) (CommitCommentReactor, error) {
	r, ok := c.(CommitCommentReactor)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureCommitCommentReactions),
			Message: fmt.Sprintf("commit comment reactions are not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return r, nil
}
