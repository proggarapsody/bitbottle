package factory

import (
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aleksey/bitbottle/api"
	"github.com/aleksey/bitbottle/internal/bbinstance"
	"github.com/aleksey/bitbottle/internal/config"
	"github.com/aleksey/bitbottle/internal/keyring"
	"github.com/aleksey/bitbottle/internal/run"
	"github.com/aleksey/bitbottle/pkg/cmdutil"
	"github.com/aleksey/bitbottle/pkg/iostreams"
)

// Factory is the single dependency container threaded through every command.
// Commands receive it via their constructor.
type Factory struct {
	IOStreams   *iostreams.IOStreams
	Config      func() (*config.Config, error)
	HttpClient  func(hostname string) (*api.Client, error)
	GitRunner   func() run.Runner
	Keyring     keyring.Keyring
	Browser     cmdutil.BrowserLauncher
	Editor      cmdutil.EditorLauncher
	BaseURL     func(hostname string) string
	Now         func() time.Time
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
		HttpClient: func(hostname string) (*api.Client, error) {
			// Ensure config is loaded before reading host settings.
			if err := loadConfig(); err != nil {
				return nil, err
			}
			hostCfg, _ := cfg.Get(hostname)
			transport := http.DefaultTransport.(*http.Transport).Clone()
			if hostCfg.SkipTLSVerify {
				transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
			}
			hc := &http.Client{Transport: transport}
			auth := api.AuthConfig{Token: hostCfg.OAuthToken}
			return api.NewClient(hc, baseURL(hostname), auth), nil
		},
		GitRunner:  func() run.Runner { return &run.SystemRunner{} },
		Keyring:    &keyring.OSKeyring{},
		Browser:    &cmdutil.SystemBrowser{},
		Editor:     &cmdutil.SystemEditor{},
		BaseURL:    baseURL,
		Now:        time.Now,
	}
}

func configHomeDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}
