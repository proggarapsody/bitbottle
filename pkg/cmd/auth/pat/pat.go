// Package pat implements the `auth pat` subgroup for managing Bitbucket
// Server/DC personal access tokens.
package pat

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPAT returns the `auth pat` subgroup command.
func NewCmdPAT(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pat",
		Short: "Manage personal access tokens for a Bitbucket host",
	}
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdRevoke(f))
	return cmd
}

// resolveHostAndUser picks the host + userSlug an auth pat command should
// operate on. If hostnameFlag is non-empty it must already be in the config.
// Returns the hostname and the userSlug stored in the config for that host.
func resolveHostAndUser(f *factory.Factory, hostnameFlag string) (string, string, error) {
	cfg, err := f.Config()
	if err != nil {
		return "", "", err
	}

	var hostname string
	if hostnameFlag != "" {
		if _, ok := cfg.Get(hostnameFlag); !ok {
			return "", "", fmt.Errorf("not logged into %s", hostnameFlag)
		}
		hostname = hostnameFlag
	} else {
		hosts := cfg.Hosts()
		switch len(hosts) {
		case 0:
			return "", "", fmt.Errorf("not authenticated; run `bitbottle auth login` first")
		case 1:
			hostname = hosts[0]
		default:
			return "", "", fmt.Errorf("multiple hosts configured; use --hostname to specify one")
		}
	}

	hc, _ := cfg.Get(hostname)
	userSlug := hc.User
	if userSlug == "" {
		return "", "", fmt.Errorf("no username stored for host %s; re-run `bitbottle auth login`", hostname)
	}
	return hostname, userSlug, nil
}
