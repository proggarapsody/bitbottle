package factory

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
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/internal/keyring"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/internal/userconfig"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

type TestFactoryOpts struct {
	ConfigDir       string
	InitialConfig   string
	HTTPClient      HTTPClient
	BaseURL         func(hostname string) string
	GitRunner       run.Runner
	Keyring         keyring.Keyring
	Browser         cmdutil.BrowserLauncher
	Editor          cmdutil.EditorLauncher
	IOStreams       *iostreams.IOStreams
	Hostname        string
	Now             func() time.Time
	BackendOverride backend.Client
	BackendType     string
}

func stubServerPATURLProber(hostname, username string, _ bool) string {
	return bbinstance.PATManageURL(hostname, username)
}

func NewTestFactory(t *testing.T, opts TestFactoryOpts) (*Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	configDir := opts.ConfigDir
	if configDir == "" {
		configDir = t.TempDir()
	}
	if opts.InitialConfig != "" {
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatalf("testfactory: mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(configDir, "hosts.yml"), []byte(opts.InitialConfig), 0o600); err != nil {
			t.Fatalf("testfactory: write config: %v", err)
		}
	}

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	ios := opts.IOStreams
	if ios == nil {
		ios = &iostreams.IOStreams{
			In:          io.NopCloser(strings.NewReader("")),
			Out:         out,
			ErrOut:      errOut,
			IsStdoutTTY: func() bool { return false },
			IsStderrTTY: func() bool { return false },
		}
	}

	gitRunner := opts.GitRunner
	if gitRunner == nil {
		gitRunner = testhelpers.NewFakeRunner()
	}

	kr := opts.Keyring
	if kr == nil {
		kr = testhelpers.NewFakeKeyring()
	}

	browser := opts.Browser
	if browser == nil {
		browser = &testhelpers.FakeBrowserLauncher{}
	}

	editor := opts.Editor
	if editor == nil {
		editor = &testhelpers.FakeEditorLauncher{}
	}

	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &noopHTTPClient{}
	}

	baseURL := opts.BaseURL
	if baseURL == nil {
		baseURL = func(h string) string {
			return "https://" + h + "/rest/api/1.0"
		}
	}

	now := opts.Now
	if now == nil {
		now = time.Now
	}

	cfg := config.New(configDir)
	userCfg := userconfig.New(configDir)
	aliasStore := aliases.New(configDir)

	f := &Factory{
		IOStreams: ios,
		Config: func() (*config.Config, error) {
			if err := cfg.Load(); err != nil && !os.IsNotExist(err) {
				return nil, err
			}
			return cfg, nil
		},
		Backend: func(hostname string) (backend.Client, error) {
			if opts.BackendOverride != nil {
				return opts.BackendOverride, nil
			}

			var hostCfg config.HostConfig
			if err := cfg.Load(); err == nil || os.IsNotExist(err) {
				hostCfg, _ = cfg.Get(hostname)
			}

			effectiveBackendType := opts.BackendType
			if effectiveBackendType == "" {
				effectiveBackendType = hostCfg.BackendType
			}

			if bbinstance.IsCloud(hostname, effectiveBackendType) {
				return cloud.NewClient(httpClient, baseURL(hostname), "test-token", ""), nil
			}
			return server.NewClient(httpClient, baseURL(hostname), "test-token", ""), nil
		},
		BackendWithOptions: func(hostname string, bOpts backend.Options) (backend.Client, error) {
			if opts.BackendOverride != nil {
				return opts.BackendOverride, nil
			}
			token := bOpts.Token
			if token == "" {
				token = "test-token"
			}
			if bbinstance.IsCloud(hostname, opts.BackendType) {
				return cloud.NewClient(httpClient, baseURL(hostname), token, ""), nil
			}
			return server.NewClient(httpClient, baseURL(hostname), token, ""), nil
		},
		HTTPClient: func(_ string) (HTTPClient, error) {
			return httpClient, nil
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
		GitRunner:          func() run.Runner { return gitRunner },
		Keyring:            kr,
		Browser:            browser,
		Editor:             editor,
		BaseURL:            baseURL,
		Now:                now,
		ServerPATURLProber: stubServerPATURLProber,
	}
	f.BaseRepo = DefaultBaseRepo(gitRunner, f.Config)
	return f, out, errOut
}

type noopHTTPClient struct{}

func (n *noopHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(`{"errors":[{"message":"no stub configured"}]}`)),
		Header:     make(http.Header),
	}, nil
}
