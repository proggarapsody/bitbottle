package backend

// defaultBranchListLimit caps how many branches the resolver fetches. 100 is
// large enough that any real repository's IsDefault entry sits in the first
// page, while keeping the request bounded.
const defaultBranchListLimit = 100

// DefaultBranch resolves a repository's default branch by inspecting the
// IsDefault flag on its branch list. It returns "" with no error when no
// branch in the list is marked default (e.g. on Bitbucket Cloud, whose
// branch endpoint omits IsDefault), so callers can fall back to local
// git state or another heuristic.
//
// Errors from the underlying ListBranches call are propagated unchanged so
// no command silently masks them with a literal "main".
func DefaultBranch(bl BranchLister, ns, slug string) (string, error) {
	branches, err := bl.ListBranches(ns, slug, defaultBranchListLimit)
	if err != nil {
		return "", err
	}
	for _, b := range branches {
		if b.IsDefault {
			return b.Name, nil
		}
	}
	return "", nil
}
