package factory

import (
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
	"github.com/proggarapsody/bitbottle/api/server"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/internal/keyring"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// Factory is the single dependency container threaded through every command.
// Commands receive it via their constructor.
type Factory struct {
	IOStreams          *iostreams.IOStreams
	Config             func() (*config.Config, error)
	Backend            func(hostname string) (backend.Client, error)
	BackendWithOptions func(hostname string, opts backend.Options) (backend.Client, error)
	GitRunner          func() run.Runner
	Keyring            keyring.Keyring
	Browser            cmdutil.BrowserLauncher
	Editor             cmdutil.EditorLauncher
	BaseURL            func(hostname string) string
	Now                func() time.Time
}

// New constructs a Factory wired with live dependencies.
func New() *Factory {
	configDir := filepath.Join(configHomeDir(), "bitbottle")
	cfg := config.New(configDir)

	baseURL := func(hostname string) string {
		return bbinstance.RESTBase(hostname)
	}

	loadConfig := func() error {
		if err := cfg.Load(); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	return &Factory{
		IOStreams: iostreams.System(),
		Config: func() (*config.Config, error) {
			if err := loadConfig(); err != nil {
				return nil, err
			}
			return cfg, nil
		},
		Backend: func(hostname string) (backend.Client, error) {
			if err := loadConfig(); err != nil {
				return nil, err
			}
			hostCfg, _ := cfg.Get(hostname)
			hc := newHTTPClient(hostCfg.SkipTLSVerify)
			return newBackendClient(hc, hostname, hostCfg, baseURL), nil
		},
		BackendWithOptions: func(hostname string, opts backend.Options) (backend.Client, error) {
			if err := loadConfig(); err != nil {
				return nil, err
			}
			hostCfg, _ := cfg.Get(hostname)
			// opts fields override the stored config values.
			if opts.Token != "" {
				hostCfg.OAuthToken = opts.Token
			}
			if opts.SkipTLSVerify {
				hostCfg.SkipTLSVerify = true
			}
			hc := newHTTPClient(hostCfg.SkipTLSVerify)
			return newBackendClient(hc, hostname, hostCfg, baseURL), nil
		},
		GitRunner: func() run.Runner { return &run.SystemRunner{} },
		Keyring:   &keyring.OSKeyring{},
		Browser:   &cmdutil.SystemBrowser{},
		Editor:    &cmdutil.SystemEditor{},
		BaseURL:   baseURL,
		Now:       time.Now,
	}
}

func configHomeDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}

// newHTTPClient returns an *http.Client configured with a clone of the default
// transport. If skipTLSVerify is true, TLS certificate verification is
// disabled (for self-signed DC instances).
func newHTTPClient(skipTLSVerify bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if skipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	return &http.Client{Transport: transport}
}

// newBackendClient selects and constructs the backend.Client implementation
// appropriate for hostname (Cloud vs. Data Center).
func newBackendClient(hc *http.Client, hostname string, hostCfg config.HostConfig, dcBaseURL func(string) string) backend.Client {
	if bbinstance.IsCloud(hostname, hostCfg.BackendType) {
		return cloud.NewClient(hc, bbinstance.CloudRESTBase(), hostCfg.OAuthToken, hostCfg.User)
	}
	return server.NewClient(hc, dcBaseURL(hostname), hostCfg.OAuthToken, hostCfg.User)
}
