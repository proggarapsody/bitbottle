// Package repoarg holds the single repository-reference parser shared by
// every CLI command. Before this package, ref parsing was inconsistent:
// `repo view` accepted only PROJECT/REPO (via bbrepo.Parse) while
// `branch list` accepted HOST/PROJECT/REPO (via factory.ResolveTarget's
// private parseTargetArg). ParseRef collapses both to one rule so a 3-part
// HOST/PROJECT/REPO is accepted everywhere a repo positional is accepted.
package repoarg

import (
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
)

// ParseRef parses a positional repository reference.
//
// Accepted forms:
//
//	PROJECT/REPO         → RepoRef{Project, Slug}            (host inferred later)
//	HOST/PROJECT/REPO    → RepoRef{Host, Project, Slug}
//
// A bare single token (no slash) is rejected: it is ambiguous with a
// branch/tag/PR name and must never be silently treated as a repo. Callers
// that want a default repo when no positional is given should fall back to
// f.BaseRepo() rather than passing an empty or bare string here.
func ParseRef(s string) (bbrepo.RepoRef, error) {
	if s == "" {
		return bbrepo.RepoRef{}, fmt.Errorf("empty repo ref")
	}
	parts := strings.Split(s, "/")
	switch len(parts) {
	case 2:
		return bbrepo.RepoRef{Project: parts[0], Slug: parts[1]}, nil
	case 3:
		return bbrepo.RepoRef{Host: parts[0], Project: parts[1], Slug: parts[2]}, nil
	default:
		return bbrepo.RepoRef{}, fmt.Errorf("invalid repo %q: expected [HOST/]PROJECT/REPO", s)
	}
}
