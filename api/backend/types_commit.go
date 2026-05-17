package backend

import "time"

// Commit is the domain representation of a single repository commit.
type Commit struct {
	Hash      string
	Message   string // subject line only (first line of commit message)
	Author    User
	Timestamp time.Time
	WebURL    string
}

// CommitStatus is a build / CI status reported against a commit hash.
type CommitStatus struct {
	Key         string
	State       string // SUCCESSFUL | FAILED | INPROGRESS | STOPPED
	Name        string
	Description string
	URL         string
}

// CommitStatusInput carries the parameters for reporting a build status.
type CommitStatusInput struct {
	Key         string
	State       string // SUCCESSFUL | FAILED | INPROGRESS | STOPPED
	Name        string
	URL         string
	Description string
}

// CommitComment is a comment attached to a specific commit.
// Reactions is only populated when explicitly requested (e.g. --reactions flag
// or include_reactions MCP parameter). Server/DC only.
type CommitComment struct {
	ID        int
	Author    User
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Reactions []CommentReaction // only populated when explicitly requested; Server/DC only
}

// AddCommitCommentInput carries parameters for creating a commit comment.
type AddCommitCommentInput struct {
	Body string
}
