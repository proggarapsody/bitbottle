package factory

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
	"github.com/proggarapsody/bitbottle/api/server"
	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/internal/aliases"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/internal/keyring"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/internal/userconfig"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Factory struct {
	IOStreams          *iostreams.IOStreams
	Config             func() (*config.Config, error)
	Backend            func(hostname string) (backend.Client, error)
	BackendWithOptions func(hostname string, opts backend.Options) (backend.Client, error)
	HTTPClient         func(hostname string) (HTTPClient, error)
	UserConfig         func() (*userconfig.Config, error)
	Aliases            func() (*aliases.Store, error)
	GitRunner          func() run.Runner
	Keyring            keyring.Keyring
	Browser            cmdutil.BrowserLauncher
	Editor             cmdutil.EditorLauncher
	BaseURL            func(hostname string) string
	BaseRepo           func() (bbrepo.RepoRef, error)
	Now                func() time.Time
	// ServerPATURLProber resolves the PAT management URL for Bitbucket Server/DC
	// by probing which URL format the instance accepts. Injected here so tests
	// can stub it without real network calls. nil → default HEAD probe.
	ServerPATURLProber func(hostname, username string, skipTLS bool) string
}

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

	configFn := func() (*config.Config, error) {
		if err := loadConfig(); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	userCfg := userconfig.New(configDir)
	userConfigFn := func() (*userconfig.Config, error) {
		if err := userCfg.Load(); err != nil {
			return nil, err
		}
		return userCfg, nil
	}

	aliasStore := aliases.New(configDir)
	aliasesFn := func() (*aliases.Store, error) {
		if err := aliasStore.Load(); err != nil {
			return nil, err
		}
		return aliasStore, nil
	}
	gitRunner := func() run.Runner { return &run.SystemRunner{} }

	return &Factory{
		IOStreams: iostreams.System(),
		Config:    configFn,
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
			if opts.Token != "" {
				hostCfg.OAuthToken = opts.Token
			}
			if opts.SkipTLSVerify {
				hostCfg.SkipTLSVerify = true
			}
			if opts.Email != "" {
				hostCfg.AuthUser = opts.Email
			}
			if opts.Username != "" {
				hostCfg.User = opts.Username
				hostCfg.AuthUser = opts.Username
			}
			hc := newHTTPClient(hostCfg.SkipTLSVerify)
			return newBackendClient(hc, hostname, hostCfg, baseURL), nil
		},
		HTTPClient: func(hostname string) (HTTPClient, error) {
			if err := loadConfig(); err != nil {
				return nil, err
			}
			hostCfg, _ := cfg.Get(hostname)
			return newHTTPClient(hostCfg.SkipTLSVerify), nil
		},
		UserConfig: userConfigFn,
		Aliases:    aliasesFn,
		GitRunner:  gitRunner,
		Keyring:    &keyring.OSKeyring{},
		Browser:    &cmdutil.SystemBrowser{},
		Editor:     &cmdutil.SystemEditor{},
		BaseURL:    baseURL,
		BaseRepo:   DefaultBaseRepo(gitRunner(), configFn),
		Now:        time.Now,
	}
}

func DefaultBaseRepo(runner run.Runner, cfg func() (*config.Config, error)) func() (bbrepo.RepoRef, error) {
	return func() (bbrepo.RepoRef, error) {
		c, err := cfg()
		if err != nil {
			return bbrepo.RepoRef{}, err
		}
		hosts := c.Hosts()

		g := git.New(runner)
		if ref, ok := readPinnedDefaultRepo(g); ok {
			return ref, nil
		}

		remoteURL, gerr := g.RemoteURL("origin")
		if gerr != nil {
			if len(hosts) == 0 {
				return bbrepo.RepoRef{}, fmt.Errorf("not authenticated; run `bitbottle auth login` first")
			}
			return bbrepo.RepoRef{}, fmt.Errorf("no git remotes found; pass [HOST/]PROJECT/REPO or run from a Bitbucket checkout")
		}

		ref, err := bbrepo.InferFromRemote(remoteURL)
		if err != nil {
			return bbrepo.RepoRef{}, err
		}
		return ref, nil
	}
}

func readPinnedDefaultRepo(g *git.Git) (bbrepo.RepoRef, bool) {
	host, _ := g.GetConfig("bitbottle.host")
	project, _ := g.GetConfig("bitbottle.project")
	slug, _ := g.GetConfig("bitbottle.slug")
	if host == "" || project == "" || slug == "" {
		return bbrepo.RepoRef{}, false
	}
	return bbrepo.RepoRef{Host: host, Project: project, Slug: slug}, true
}

func configHomeDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}

func newHTTPClient(skipTLSVerify bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if skipTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	return &http.Client{Transport: transport}
}

func newBackendClient(hc *http.Client, hostname string, hostCfg config.HostConfig, dcBaseURL func(string) string) backend.Client {
	authUser := hostCfg.AuthUser
	if authUser == "" {
		authUser = hostCfg.User // backward-compat: older configs have no AuthUser
	}
	if bbinstance.IsCloud(hostname, hostCfg.BackendType) {
		return cloud.NewClient(hc, bbinstance.CloudRESTBase(), hostCfg.OAuthToken, authUser)
	}
	return server.NewClient(hc, dcBaseURL(hostname), hostCfg.OAuthToken, hostCfg.User)
}
