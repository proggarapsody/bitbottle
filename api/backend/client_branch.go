package backend

// BranchLister lists branches in a repository.
type BranchLister interface {
	ListBranches(ns, slug string, limit int) ([]Branch, error)
}

// BranchCreator creates a branch.
type BranchCreator interface {
	CreateBranch(ns, slug string, in CreateBranchInput) (Branch, error)
}

// BranchDeleter deletes a branch.
type BranchDeleter interface {
	DeleteBranch(ns, slug, branch string) error
}

// BranchProtector exposes branch-restriction management on Bitbucket
// Server / Data Center. The Cloud backend has a different "branch
// restrictions" shape that's not modelled here — calls against a Cloud
// client surface ErrUnsupportedOnHost via AsBranchProtector.
type BranchProtector interface {
	ListBranchProtections(ns, slug string, limit int) ([]BranchProtection, error)
	CreateBranchProtection(ns, slug string, in CreateBranchProtectionInput) (BranchProtection, error)
	DeleteBranchProtection(ns, slug string, id int) error
}

// FeatureBranchProtect names the branch-protection capability for typed-
// error reporting.
const FeatureBranchProtect Feature = "branch-protect"

// AsBranchProtector returns the BranchProtector view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a backend
// that doesn't model branch protections (currently Cloud).
func AsBranchProtector(c Client, host string) (BranchProtector, error) {
	return requireFeature[BranchProtector](c, host, specFor(FeatureBranchProtect))
}
