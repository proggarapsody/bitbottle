package backend

// CherryPickInput carries the parameters for cherry-picking a commit.
type CherryPickInput struct {
	SourceHash   string // commit hash to cherry-pick
	TargetBranch string // destination branch name
	Message      string // optional commit message override (empty = use original)
}

// CommitCherryPicker can cherry-pick a commit onto a target branch.
type CommitCherryPicker interface {
	CherryPickCommit(ns, slug string, in CherryPickInput) (Commit, error)
}

// FeatureCherryPick names the cherry-pick capability.
const FeatureCherryPick Feature = "CommitCherryPicker"

// AsCommitCherryPicker returns the CommitCherryPicker view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a backend
// that doesn't implement cherry-pick operations.
func AsCommitCherryPicker(c Client, host string) (CommitCherryPicker, error) {
	return requireFeature[CommitCherryPicker](c, host, specFor(FeatureCherryPick))
}
