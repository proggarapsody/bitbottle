package backend

// CommitCommenter manages comments on individual commits.
type CommitCommenter interface {
	ListCommitComments(ns, slug, hash string, limit int) ([]CommitComment, error)
	AddCommitComment(ns, slug, hash string, in AddCommitCommentInput) (CommitComment, error)
	EditCommitComment(ns, slug, hash string, commentID int, body string) (CommitComment, error)
	DeleteCommitComment(ns, slug, hash string, commentID int) error
}
