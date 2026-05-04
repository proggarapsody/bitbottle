package factory

import (
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
)

// ResolveTarget returns the repository the command should act on.
//
// Resolution order:
//  1. If hostnameFlag is set, it overrides whichever Host was resolved.
//  2. If args contains exactly one element, it is parsed as
//     [HOST/]PROJECT/REPO. (Implemented in a later cycle.)
//  3. Otherwise f.BaseRepo() is consulted — which already honours
//     -R/--repo, BB_REPO, pinned git config, and the origin remote.
//
// This is the single resolver every command should use. It replaces:
//   - factory.Factory.ResolveRef (REQUIRED positional arg, ignored BaseRepo)
//   - pr.resolveRepoRef          (PR-package local duplicate)
//   - pr.resolvePRTarget         (now layered on top of this)
//
// Effect: -R works everywhere, not just `pr list`. Commands that today
// require PROJECT/REPO can accept zero args and fall back to BaseRepo.
func ResolveTarget(f *Factory, args []string, hostnameFlag string) (bbrepo.RepoRef, error) {
	if len(args) == 0 {
		ref, err := f.BaseRepo()
		if err != nil {
			return bbrepo.RepoRef{}, err
		}
		if hostnameFlag != "" {
			ref.Host = hostnameFlag
		}
		return ref, nil
	}

	// User passed an explicit target — parse [HOST/]PROJECT/REPO.
	ref, err := parseTargetArg(args[0])
	if err != nil {
		return bbrepo.RepoRef{}, err
	}
	if hostnameFlag != "" {
		ref.Host = hostnameFlag
		return ref, nil
	}
	if ref.Host == "" {
		// Bare PROJECT/REPO — infer host iff exactly one is configured.
		// Multi-host ambiguity is surfaced rather than guessed at.
		if err := inferHost(f, &ref); err != nil {
			return bbrepo.RepoRef{}, err
		}
	}
	return ref, nil
}

func inferHost(f *Factory, ref *bbrepo.RepoRef) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}
	hosts := cfg.Hosts()
	switch len(hosts) {
	case 0:
		return fmt.Errorf("not authenticated; run `bitbottle auth login` first")
	case 1:
		ref.Host = hosts[0]
		return nil
	default:
		return fmt.Errorf("multiple hosts configured; specify HOST/PROJECT/REPO or use --hostname")
	}
}

// parseTargetArg accepts "HOST/PROJECT/REPO" or "PROJECT/REPO".
// PROJECT-only-with-host inference happens in U3.
func parseTargetArg(s string) (bbrepo.RepoRef, error) {
	parts := strings.Split(s, "/")
	switch len(parts) {
	case 2:
		return bbrepo.RepoRef{Project: parts[0], Slug: parts[1]}, nil
	case 3:
		return bbrepo.RepoRef{Host: parts[0], Project: parts[1], Slug: parts[2]}, nil
	default:
		return bbrepo.RepoRef{}, fmt.Errorf("invalid target %q: expected [HOST/]PROJECT/REPO", s)
	}
}
