package backend

import "time"

// Workspace is a Bitbucket Cloud workspace (the top-level ownership unit
// containing repositories and projects). Bitbucket Server / Data Center has
// no analogous concept — projects sit directly under the instance — so
// Workspace is Cloud-only and surfaced via the WorkspaceClient optional
// interface.
type Workspace struct {
	Slug   string
	Name   string
	UUID   string
	WebURL string
}

// WorkspaceMember is a member of a Bitbucket Cloud workspace.
type WorkspaceMember struct {
	User      User
	Workspace string // workspace slug
}

// Project is a Bitbucket Cloud project (a logical group of repositories
// inside a workspace). The naming clashes with Server/DC's "project" — which
// is the namespace itself — but Cloud's project sits one level deeper.
// Listed via WorkspaceClient.ListProjects(workspace).
type Project struct {
	Key    string
	Name   string
	UUID   string
	WebURL string
}

// Issue is a Bitbucket Cloud issue. Issues are a Cloud-only feature gated
// by per-repository "issue tracker" enablement; the API returns 404 on
// repositories where the tracker is disabled. We surface that as the
// adapter's standard ErrNotFound.
//
// Assignee is a pointer so the zero value cleanly distinguishes
// "unassigned" from "assigned to user with empty username". Reporter is
// always present on Cloud and uses the value type.
type Issue struct {
	ID        int
	Title     string
	State     string // new, open, on hold, resolved, duplicate, invalid, wontfix, closed
	Kind      string // bug, enhancement, proposal, task
	Priority  string // trivial, minor, major, critical, blocker
	Reporter  User
	Assignee  *User // nil when unassigned
	CreatedOn time.Time
	UpdatedOn time.Time
	WebURL    string
	Content   string // raw markdown body
}

// CreateIssueInput carries the parameters for opening a new issue. Bitbucket
// Cloud applies sane defaults (kind=bug, priority=major) when fields are
// empty, so callers can omit everything but Title.
type CreateIssueInput struct {
	Title    string
	Content  string
	Kind     string
	Priority string
}

// UpdateIssueInput carries the parameters for changing an issue. Empty
// strings mean "no change". `issue close` sets State="closed" and leaves
// the rest untouched. Assignee is the Bitbucket Cloud username to assign to;
// set to the special sentinel AssigneeNone to explicitly clear the assignee.
type UpdateIssueInput struct {
	Title    string
	Content  string
	State    string
	Kind     string
	Priority string
	Assignee string // "" = no change; AssigneeNone = clear
}

// AssigneeNone is a sentinel value for UpdateIssueInput.Assignee that
// signals "unassign" (set assignee to null on the wire).
const AssigneeNone = "__none__"

// IssueComment is a comment on a Bitbucket Cloud issue.
type IssueComment struct {
	ID        int
	Author    User
	Content   string
	CreatedOn time.Time
	UpdatedOn time.Time
}

// IssueAttachment is a file attached to a Bitbucket Cloud issue.
type IssueAttachment struct {
	Name     string
	Size     int64
	MIMEType string
	Links    struct{ Self string }
}

// Webhook is the domain representation of a repository webhook.
// Both Bitbucket Cloud and Server/DC expose a similar shape — a remote URL,
// a list of subscribed events, and an active flag. ID is the backend's
// stable identifier (UUID on Cloud, integer-as-string on Server/DC).
type Webhook struct {
	ID     string
	URL    string
	Events []string
	Active bool
}

// CreateWebhookInput carries the parameters for creating a webhook.
// Secret is write-only — neither backend returns it on read.
type CreateWebhookInput struct {
	URL    string
	Events []string
	Active bool
	Secret string
}
