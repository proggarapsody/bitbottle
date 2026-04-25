package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdAuth(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with a Bitbucket host",
	}
	cmd.AddCommand(NewCmdAuthLogin(f))
	cmd.AddCommand(NewCmdAuthLogout(f))
	cmd.AddCommand(NewCmdAuthStatus(f))
	return cmd
}

// resolveAuthHostname picks the host an auth command should operate on.
//
// Resolution order:
//  1. If hostnameFlag is non-empty, the host must already be present in the
//     config (so we surface a "not logged into X" error rather than acting on
//     an unknown host).
//  2. Otherwise, fall back to the single configured host.
//  3. Zero or multiple configured hosts produce a deterministic error.
//
// Returning the loaded config alongside the host avoids each caller having to
// re-load it separately.
func resolveAuthHostname(f *factory.Factory, hostnameFlag string) (*config.Config, string, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, "", err
	}

	if hostnameFlag != "" {
		if _, ok := cfg.Get(hostnameFlag); !ok {
			return nil, "", fmt.Errorf("not logged into %s", hostnameFlag)
		}
		return cfg, hostnameFlag, nil
	}

	hosts := cfg.Hosts()
	switch len(hosts) {
	case 0:
		return nil, "", fmt.Errorf("not authenticated; run `bitbottle auth login` first")
	case 1:
		return cfg, hosts[0], nil
	default:
		return nil, "", fmt.Errorf("multiple hosts configured; use --hostname to specify one")
	}
}
