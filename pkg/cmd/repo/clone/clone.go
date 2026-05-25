package clone

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdClone(f *factory.Factory) *cobra.Command {
	var hostname string
	var useSSH bool
	var useHTTPS bool

	cmd := &cobra.Command{
		Use:   "clone PROJECT/REPO [DIR]",
		Short: "Clone a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}
			ref.Host = host

			cfg, err := f.Config()
			if err != nil {
				return err
			}
			hostCfg, _ := cfg.Get(host)

			// Determine target directory.
			var dir string
			if len(args) == 2 {
				dir = args[1]
			} else {
				dir = ref.Slug
			}

			// Try to resolve the clone URL via API.
			cloneURL := resolveCloneURL(f, host, ref, hostCfg, useSSH, useHTTPS)

			g := git.New(f.GitRunner())
			if err := g.Clone(cloneURL, dir); err != nil {
				return err
			}

			// Write bitbottle git config keys into the cloned repo.
			_ = g.SetConfigInDir(dir, "bitbottle.host", host)
			_ = g.SetConfigInDir(dir, "bitbottle.project", ref.Project)
			_ = g.SetConfigInDir(dir, "bitbottle.slug", ref.Slug)

			fmt.Fprintf(f.IOStreams.Out, "Cloned to %s\n", dir)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().BoolVar(&useSSH, "ssh", false, "Clone via SSH")
	cmd.Flags().BoolVar(&useHTTPS, "https", false, "Clone via HTTPS")
	return cmd
}

// resolveCloneURL tries the API first; falls back to buildCloneURL.
func resolveCloneURL(f *factory.Factory, host string, ref bbrepo.RepoRef, hostCfg config.HostConfig, useSSH, useHTTPS bool) string {
	client, err := f.Backend(host)
	if err == nil {
		repo, err := client.GetRepo(ref.Project, ref.Slug)
		if err == nil && len(repo.CloneURLs) > 0 {
			if url := pickCloneURL(repo.CloneURLs, hostCfg, useSSH, useHTTPS); url != "" {
				return url
			}
		}
	}
	return buildCloneURL(host, ref, hostCfg)
}

// pickCloneURL selects a URL from the API list based on flags and config.
func pickCloneURL(urls []backend.CloneURL, hostCfg config.HostConfig, useSSH, useHTTPS bool) string {
	// Determine preferred protocol.
	preferSSH := useSSH || (!useHTTPS && hostCfg.GitProtocol != "https" && hostCfg.GitProtocol != "http")

	if preferSSH {
		for _, u := range urls {
			if u.Name == "ssh" {
				return u.URL
			}
		}
	}
	// HTTPS: accept "https" or "http".
	for _, u := range urls {
		if u.Name == "https" || u.Name == "http" {
			return u.URL
		}
	}
	// Fallback: return first.
	if len(urls) > 0 {
		return urls[0].URL
	}
	return ""
}

func buildCloneURL(host string, ref bbrepo.RepoRef, hostCfg config.HostConfig) string {
	protocol := hostCfg.GitProtocol
	if protocol == "" {
		protocol = "ssh"
	}
	isCloud := bbinstance.IsCloud(host, hostCfg.BackendType)

	if protocol == "ssh" {
		if isCloud {
			return bbinstance.CloudSSHURL(ref.Project, ref.Slug)
		}
		return fmt.Sprintf("ssh://git@%s/%s/%s.git", host, ref.Project, ref.Slug)
	}

	if isCloud {
		return bbinstance.CloudHTTPSURL(ref.Project, ref.Slug)
	}
	return bbinstance.HTTPSURL(host, ref.Project, ref.Slug)
}

func resolveHostname(f *factory.Factory, flag string) (string, error) {
	if flag != "" {
		return flag, nil
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
		return "", fmt.Errorf("multiple hosts configured; use --hostname to specify one")
	}
}
