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

// ResolveHost returns the host the caller should target.
//
// Rules:
//  1. If hostname is non-empty, it wins — no config lookup is performed.
//     Callers that already know which host they want (e.g. a --hostname
//     flag, or an MCP tool's `hostname` arg) shouldn't pay for IO.
//  2. Otherwise consult config:
//     - exactly one host configured  → use it
//     - zero hosts                   → "not authenticated" error
//     - multiple hosts               → "specify hostname" error
//
// This is the single host-inference rule shared by both surfaces:
// the CLI's ResolveTarget (when args contain a bare PROJECT/REPO) and
// the MCP handlers (which take hostname as a tool arg, never positional).
func ResolveHost(f *Factory, hostname string) (string, error) {
	if hostname != "" {
		return hostname, nil
	}
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	hosts := cfg.Hosts()
	switch len(hosts) {
	case 0:
		return "", fmt.Errorf("not authenticated; run `bitbottle auth login` first")
	case 1:
		return hosts[0], nil
	default:
		return "", fmt.Errorf("multiple hosts configured; specify hostname")
	}
}

// inferHost is the bare-PROJECT/REPO branch of ResolveTarget. The
// host-discovery rule lives in ResolveHost; this wrapper exists only
// to surface the *target-specific* multi-host error message that
// guides users toward HOST/PROJECT/REPO rather than a bare --hostname.
func inferHost(f *Factory, ref *bbrepo.RepoRef) error {
	host, err := ResolveHost(f, "")
	if err != nil {
		// Rewrite the multi-host message so it fits the positional-arg
		// context (the user just typed PROJECT/REPO; tell them to add
		// the host or pass --hostname).
		if strings.Contains(err.Error(), "multiple hosts") {
			return fmt.Errorf("multiple hosts configured; specify HOST/PROJECT/REPO or use --hostname")
		}
		return err
	}
	ref.Host = host
	return nil
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
