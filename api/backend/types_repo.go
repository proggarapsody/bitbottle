package backend

// CloneURL is a single clone-protocol entry returned by the Bitbucket repo API.
type CloneURL struct {
	Name string // "ssh", "https", or "http"
	URL  string
}

// Repository is the domain representation of a Bitbucket repository.
//
// ID is the backend's numeric identifier (Server / Data Center). Cloud
// uses opaque UUIDs surfaced through other channels; for Cloud repos
// this field stays zero. ID is exposed so the BBS default-reviewers
// endpoint (which requires source/target repo IDs as query params) has
// what it needs without an adapter-internal cache.
type Repository struct {
	Slug        string
	Name        string
	Namespace   string
	SCM         string
	WebURL      string
	Description string
	IsPrivate   bool
	ID          int
	CloneURLs   []CloneURL
}

// CreateRepoInput carries the parameters for creating a repository.
type CreateRepoInput struct {
	Name        string
	SCM         string
	Public      bool
	Description string
}

// ForkRepoInput carries the parameters for forking a Bitbucket Cloud
// repository. Workspace is required (the fork's destination workspace).
// Name is optional — when empty, Bitbucket Cloud reuses the source name.
type ForkRepoInput struct {
	Workspace string
	Name      string
}

// Branch is the domain representation of a repository branch.
type Branch struct {
	Name       string
	IsDefault  bool
	LatestHash string
}

// CreateBranchInput carries the parameters for creating a branch.
type CreateBranchInput struct {
	Name    string
	StartAt string // branch name or commit hash
}

// Tag is the domain representation of a repository tag.
type Tag struct {
	Name    string
	Hash    string
	Message string // empty for lightweight tags; first line for annotated
	WebURL  string
}

// CreateTagInput carries the parameters for creating a tag.
type CreateTagInput struct {
	Name    string
	StartAt string // branch name or commit hash
	Message string // empty = lightweight tag; non-empty = annotated tag
}

// BranchProtection is a single branch-restriction record on Bitbucket
// Server / Data Center. Cloud has a similar concept under a different,
// non-trivial wire shape that we don't model yet — BranchProtection is
// surfaced only via the BranchProtector optional interface.
//
// Type values match BBS exactly (lowercased, dash-separated):
//
//	read-only           — disallow all writes
//	no-deletes          — disallow branch deletion
//	fast-forward-only   — disallow non-fast-forward writes
//	pull-request-only   — only PR merges may write
//
// MatcherKind is the BBS matcher type id ("BRANCH", "PATTERN",
// "MODEL_BRANCH", "MODEL_CATEGORY") and MatcherID is the corresponding
// value (a branch name, glob, or model id). Users / Groups list slugs
// that are exempted from the restriction.
type BranchProtection struct {
	ID          int
	Type        string
	MatcherID   string
	MatcherKind string
	Users       []string
	Groups      []string
}

// CreateBranchProtectionInput carries the parameters for adding a
// branch-restriction. All fields except Type and MatcherID are optional.
// MatcherKind defaults to "BRANCH" when empty (the most common case).
type CreateBranchProtectionInput struct {
	Type        string
	MatcherID   string
	MatcherKind string
	Users       []string
	Groups      []string
}

// BranchRule is the domain representation of a Bitbucket Cloud branch restriction rule.
type BranchRule struct {
	ID      int    `json:"id"`
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
	Value   int    `json:"value,omitempty"`
}

// BranchRuleInput carries the parameters for adding a branch restriction rule.
type BranchRuleInput struct {
	Kind    string
	Pattern string
	Value   int
}

// UpdateBranchRuleInput carries the patch parameters for updating a branch restriction rule.
// A nil pointer means "no change"; non-nil overwrites the current value.
// For Users and Groups, a non-nil empty slice clears the existing entries.
type UpdateBranchRuleInput struct {
	Pattern *string
	Users   *[]string
	Groups  *[]string
	Value   *int
}

// DeployKey is the domain representation of a repository deploy key (SSH public key).
type DeployKey struct {
	ID       int    `json:"id"`
	Label    string `json:"label"`
	Key      string `json:"key"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}

// DeployKeyInput carries the parameters for adding a deploy key.
type DeployKeyInput struct {
	Label      string
	Key        string
	Permission string // "read" | "read_write"; empty → omit from wire (default: read). Cloud only.
}

// DiffStat is the domain representation of a repository diff summary.
type DiffStat struct {
	FilesChanged int
	Additions    int
	Deletions    int
	Files        []DiffStatEntry
}

// DiffStatEntry is the domain representation of a single file's diff summary.
type DiffStatEntry struct {
	Path      string
	Status    string // "added", "modified", "deleted", "renamed"
	Additions int
	Deletions int
}

// TreeEntry is a single child of a directory listing under a ref. Type is
// the normalised "file" or "dir" — adapters fold their backend-specific
// vocabularies (Cloud "commit_file"/"commit_directory", Server
// "FILE"/"DIRECTORY"/"SUBMODULE") to one of these two values. Submodules
// surface as "dir" with the submodule pointer in Hash so renderers treat
// them as recursable.
//
// Path is the full repo-relative path (including the parent prefix), so a
// renderer doesn't need to know what path was requested to display the
// listing. Size is 0 for directories.
type TreeEntry struct {
	Path string
	Type string
	Size int64
	Hash string
}

// RepoLabel is a label that can be attached to a Bitbucket repository.
type RepoLabel struct {
	ID    int
	Name  string
	Color string
}

// CreateRepoLabelInput carries the parameters for creating a repository label.
type CreateRepoLabelInput struct {
	Name  string
	Color string
}

// UpdateRepoLabelInput carries the parameters for updating a repository label.
type UpdateRepoLabelInput struct {
	Name  string
	Color string
}
