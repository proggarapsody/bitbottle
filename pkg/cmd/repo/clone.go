package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoClone(f *factory.Factory) *cobra.Command {
	var hostname string

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

			cloneURL := buildCloneURL(host, ref, hostCfg)

			var dir string
			if len(args) == 2 {
				dir = args[1]
			}

			g := git.New(f.GitRunner())
			return g.Clone(cloneURL, dir)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
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
