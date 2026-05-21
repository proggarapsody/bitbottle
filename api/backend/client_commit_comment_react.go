package backend

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
	return requireFeature[CommitCommentReactor](c, host, specFor(FeatureCommitCommentReactions))
}
