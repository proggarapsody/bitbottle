// Package factorytest builds a *factory.Factory for tests with sane,
// minimal defaults. It replaces the old factory.NewTestFactory god-helper
// with a thin layer: New(t) gives you a Factory and a buffers handle;
// individual fields you care about, you override directly on the
// returned Factory. This mirrors how cli/cli's tests build factories —
// a literal with a few field assignments — instead of routing every
// test through a 13-field options struct.
package factorytest

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
	"github.com/proggarapsody/bitbottle/api/server"
	"github.com/proggarapsody/bitbottle/internal/aliases"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/internal/profiles"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/internal/userconfig"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// Opts narrows the options surface to the two pieces of setup that
// are inconvenient to do *after* New() returns: the config dir
// (because Config/UserConfig/Aliases all close over it) and the
// hosts.yml content (because it has to land on disk before the first
// f.Config() call). Everything else — IOStreams, GitRunner, Backend,
// HTTPClient, BaseURL, Now — is overridden on the returned Factory
// directly. The 11 invisible defaults of the old TestFactoryOpts are
// gone: tests that care about a field declare it.
type Opts struct {
	ConfigDir     string // optional; defaults to t.TempDir()
	InitialConfig string // optional; written to <ConfigDir>/hosts.yml
	// BackendType forces cloud-vs-server dispatch in the default
	// Backend wiring. Useful when the hostname doesn't make routing
	// obvious (e.g. testing cloud behaviour against git.example.com).
	// Empty string → infer from hostname / hosts.yml as production does.
	BackendType string
}

// New returns a Factory wired with capture buffers, fake git/keyring/
// browser/editor, a 404 HTTP stub, and a config loader pointing at a
// temp dir. The returned (out, errOut) are the *bytes.Buffer instances
// underlying f.IOStreams.Out and f.IOStreams.ErrOut.
//
// f.Backend / f.BackendWithOptions read f.HTTPClient at call time, so
// tests that swap the HTTP stub *after* New() automatically get a
// backend client wired to the new stub.
func New(t *testing.T, opts Opts) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	configDir := opts.ConfigDir
	if configDir == "" {
		configDir = t.TempDir()
	}
	if opts.InitialConfig != "" {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatalf("factorytest: mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(configDir, "hosts.yml"), []byte(opts.InitialConfig), 0o600); err != nil {
			t.Fatalf("factorytest: write config: %v", err)
		}
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := &iostreams.IOStreams{
		In:          io.NopCloser(strings.NewReader("")),
		Out:         out,
		ErrOut:      errOut,
		IsStdoutTTY: func() bool { return false },
		IsStderrTTY: func() bool { return false },
	}

	cfg := config.New(configDir)
	userCfg := userconfig.New(configDir)
	aliasStore := aliases.New(configDir)
	profileStore := profiles.New(configDir)
	gitRunner := testhelpers.NewFakeRunner()

	defaultBaseURL := func(h string) string {
		return "https://" + h + "/rest/api/1.0"
	}

	f := &factory.Factory{
		IOStreams: ios,
		Config: func() (*config.Config, error) {
			if err := cfg.Load(); err != nil && !os.IsNotExist(err) {
				return nil, err
			}
			return cfg, nil
		},
		UserConfig: func() (*userconfig.Config, error) {
			if err := userCfg.Load(); err != nil {
				return nil, err
			}
			return userCfg, nil
		},
		Aliases: func() (*aliases.Store, error) {
			if err := aliasStore.Load(); err != nil {
				return nil, err
			}
			return aliasStore, nil
		},
		Profiles: func() (*profiles.Store, error) {
			if err := profileStore.Load(); err != nil {
				return nil, err
			}
			return profileStore, nil
		},
		HTTPClient: func(_ string) (factory.HTTPClient, error) {
			return &noopHTTPClient{}, nil
		},
		GitRunner:          func() run.Runner { return gitRunner },
		Keyring:            testhelpers.NewFakeKeyring(),
		Browser:            &testhelpers.FakeBrowserLauncher{},
		Editor:             &testhelpers.FakeEditorLauncher{},
		BaseURL:            defaultBaseURL,
		Now:                time.Now,
		ServerPATURLProber: stubServerPATURLProber,
	}
	// Backend/BackendWithOptions read f.HTTPClient at call time so
	// tests that swap the HTTP stub (the common case in api/auth
	// tests) automatically get a backend client wired to the new
	// stub. BackendType from Opts wins; otherwise infer from hosts.yml.
	backendType := opts.BackendType
	f.Backend = func(hostname string) (backend.Client, error) {
		return defaultBackend(f, hostname, "", backendType)
	}
	f.BackendWithOptions = func(hostname string, bOpts backend.Options) (backend.Client, error) {
		return defaultBackend(f, hostname, bOpts.Token, backendType)
	}
	// BaseRepo resolves through f.GitRunner() at call time so tests
	// that swap the runner *after* New() (e.g. to inject "exit status
	// 128" for the no-git-repo path) see the new runner.
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return factory.DefaultBaseRepo(f.GitRunner(), f.Config)()
	}
	return f, out, errOut
}

// defaultBackend dials a real cloud/server client through f.HTTPClient
// and f.BaseURL — both of which the test may have overridden after
// New() returned. Token defaults to "test-token". explicitType wins
// over hosts.yml when non-empty.
func defaultBackend(f *factory.Factory, hostname, token, explicitType string) (backend.Client, error) {
	if token == "" {
		token = "test-token"
	}
	hc, err := f.HTTPClient(hostname)
	if err != nil {
		return nil, err
	}
	effectiveType := explicitType
	if effectiveType == "" {
		cfg, _ := f.Config()
		if cfg != nil {
			hostCfg, _ := cfg.Get(hostname)
			effectiveType = hostCfg.BackendType
		}
	}
	if bbinstance.IsCloud(hostname, effectiveType) {
		return cloud.NewClient(hc, f.BaseURL(hostname), token, ""), nil
	}
	return server.NewClient(hc, f.BaseURL(hostname), token, ""), nil
}

// StubBackend returns a Backend func suitable for f.Backend assignment
// when a test wants every host to dial the same stub client. NOTE:
// most tests should call UseBackend(f, client) instead, which also
// wires f.BackendWithOptions — auth flows go through the latter and
// silently fail otherwise.
func StubBackend(client backend.Client) func(string) (backend.Client, error) {
	return func(string) (backend.Client, error) { return client, nil }
}

// UseBackend wires both f.Backend and f.BackendWithOptions to return
// the same stub client. This matches the old NewTestFactory's
// BackendOverride behaviour, which silently routed both call paths
// to the override. Auth login/logout/refresh tests need this — they
// dial through BackendWithOptions for the initial probe.
func UseBackend(f *factory.Factory, client backend.Client) {
	f.Backend = func(string) (backend.Client, error) { return client, nil }
	f.BackendWithOptions = func(string, backend.Options) (backend.Client, error) {
		return client, nil
	}
}

// StubHTTPClient returns an HTTPClient func suitable for f.HTTPClient
// assignment when a test wants every host to dial the same stub.
func StubHTTPClient(hc factory.HTTPClient) func(string) (factory.HTTPClient, error) {
	return func(string) (factory.HTTPClient, error) { return hc, nil }
}

// stubServerPATURLProber bypasses the network probe that production
// uses to discover the right PAT URL on Bitbucket Server. Tests get
// the deterministic format from bbinstance directly.
func stubServerPATURLProber(hostname, username string, _ bool) string {
	return bbinstance.PATManageURL(hostname, username)
}

// UseProfiles wires f.Profiles to return the given store. This lets tests
// inject a pre-populated store without touching the real filesystem.
func UseProfiles(f *factory.Factory, store *profiles.Store) {
	f.Profiles = func() (*profiles.Store, error) { return store, nil }
}

// noopHTTPClient returns 404 for every request so tests that forget
// to wire a real HTTP stub fail with a clear "no stub configured"
// payload rather than a nil-deref panic.
type noopHTTPClient struct{}

func (n *noopHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"errors":[{"message":"no stub configured"}]}`)),
		Header:     make(http.Header),
	}, nil
}
