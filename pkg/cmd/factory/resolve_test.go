package factory_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// TestResolveTarget_EmptyArgs_DelegatesToBaseRepo is the tracer bullet for
// the unified resolver. Until now bitbottle had three near-identical
// resolution paths (factory.ResolveRef, pr.resolveRepoRef, pr.resolvePRTarget)
// each falling back to a different "current repo" lookup. ResolveTarget
// collapses them all to one rule: when the user hasn't typed a positional
// arg, ask f.BaseRepo() — the SAME function that -R/--repo and BB_REPO
// already mutate. Effect: -R will start working for every command, not
// just `pr list`.
func TestResolveTarget_EmptyArgs_DelegatesToBaseRepo(t *testing.T) {
	t.Parallel()

	want := bbrepo.RepoRef{Host: "git.example.com", Project: "MYPROJ", Slug: "svc"}
	calls := 0
	f := &factory.Factory{
		BaseRepo: func() (bbrepo.RepoRef, error) {
			calls++
			return want, nil
		},
	}

	got, err := factory.ResolveTarget(f, nil, "")
	require.NoError(t, err)
	assert.Equal(t, want, got, "empty args must yield the BaseRepo result")
	assert.Equal(t, 1, calls, "BaseRepo should have been consulted exactly once")
}

// TestResolveTarget_EmptyArgs_PropagatesBaseRepoError verifies the error
// returned by f.BaseRepo() reaches the caller verbatim — no friendly
// rewriting at this layer (cmdutil.ExplainError owns presentation).
func TestResolveTarget_EmptyArgs_PropagatesBaseRepoError(t *testing.T) {
	t.Parallel()

	want := errors.New("not authenticated; run `bitbottle auth login` first")
	f := &factory.Factory{
		BaseRepo: func() (bbrepo.RepoRef, error) { return bbrepo.RepoRef{}, want },
	}

	_, err := factory.ResolveTarget(f, nil, "")
	require.ErrorIs(t, err, want)
}

// TestResolveTarget_BareProjectRepo_InfersSingleHost verifies that when
// the user types only `PROJECT/REPO` and exactly one host is authenticated,
// that host is used. This preserves today's UX for users with one
// configured host — they don't have to spell it out every command.
//
// Multi-host setups error: ambiguity is not the resolver's problem to
// guess at. Users must spell the host or use --hostname.
func TestResolveTarget_BareProjectRepo_InfersSingleHost(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, writeHosts(dir, "git.example.com:\n  oauth_token: tok\n  user: alice\n"))
	cfg := config.New(dir)

	f := &factory.Factory{
		Config: func() (*config.Config, error) {
			require.NoError(t, cfg.Load())
			return cfg, nil
		},
		BaseRepo: func() (bbrepo.RepoRef, error) {
			t.Fatal("BaseRepo must NOT be called when an explicit arg is provided")
			return bbrepo.RepoRef{}, nil
		},
	}

	got, err := factory.ResolveTarget(f, []string{"MYPROJ/svc"}, "")
	require.NoError(t, err)
	assert.Equal(t, bbrepo.RepoRef{Host: "git.example.com", Project: "MYPROJ", Slug: "svc"}, got,
		"single configured host must be inferred for bare PROJECT/REPO")
}

// TestResolveTarget_BareProjectRepo_MultiHost_Errors verifies the
// ambiguity case: with multiple configured hosts and no host in the
// arg, we error rather than guess. The error guides users toward
// disambiguating with HOST/PROJECT/REPO or --hostname.
func TestResolveTarget_BareProjectRepo_MultiHost_Errors(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, writeHosts(dir,
		"git.a.com:\n  oauth_token: tok1\n  user: alice\n"+
			"git.b.com:\n  oauth_token: tok2\n  user: alice\n"))
	cfg := config.New(dir)

	f := &factory.Factory{
		Config: func() (*config.Config, error) {
			require.NoError(t, cfg.Load())
			return cfg, nil
		},
	}

	_, err := factory.ResolveTarget(f, []string{"MYPROJ/svc"}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple hosts",
		"multi-host ambiguity should surface a guiding message; got: %v", err)
}

// writeHosts is a tiny helper to seed a test config.
func writeHosts(dir, contents string) error {
	return writeFileTo(filepath.Join(dir, "hosts.yml"), contents)
}

func writeFileTo(path, contents string) error {
	return os.WriteFile(path, []byte(contents), 0o600)
}

// TestResolveTarget_HostnameFlag_OverridesEverything verifies precedence:
// the --hostname flag wins over both the host parsed from a fully-
// qualified arg and the host that BaseRepo would have returned.
// This matches what users expect from explicit overrides.
func TestResolveTarget_HostnameFlag_OverridesEverything(t *testing.T) {
	t.Parallel()

	f := &factory.Factory{
		BaseRepo: func() (bbrepo.RepoRef, error) {
			return bbrepo.RepoRef{Host: "git.from-base-repo.com", Project: "P", Slug: "S"}, nil
		},
	}

	// Case 1: empty args → BaseRepo's host gets overridden.
	got, err := factory.ResolveTarget(f, nil, "git.override.com")
	require.NoError(t, err)
	assert.Equal(t, "git.override.com", got.Host)

	// Case 2: fully-qualified arg → arg's host gets overridden.
	got, err = factory.ResolveTarget(f, []string{"git.in-arg.com/P/S"}, "git.override.com")
	require.NoError(t, err)
	assert.Equal(t, "git.override.com", got.Host)
}

// TestResolveTarget_FullyQualifiedArg verifies the explicit-target path:
// when the user passes HOST/PROJECT/REPO directly, every component is
// honored verbatim, BaseRepo is NOT consulted (the user was explicit
// about what they wanted), and no host-inference is attempted.
func TestResolveTarget_FullyQualifiedArg(t *testing.T) {
	t.Parallel()

	f := &factory.Factory{
		BaseRepo: func() (bbrepo.RepoRef, error) {
			t.Fatal("BaseRepo must NOT be called when args[0] is fully qualified")
			return bbrepo.RepoRef{}, nil
		},
	}

	got, err := factory.ResolveTarget(f, []string{"git.example.com/MYPROJ/svc"}, "")
	require.NoError(t, err)
	assert.Equal(t, bbrepo.RepoRef{Host: "git.example.com", Project: "MYPROJ", Slug: "svc"}, got)
}
