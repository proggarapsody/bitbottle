package auth

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdAuthStatus(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			hosts := cfg.Hosts()
			if len(hosts) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "You are not logged into any Bitbucket hosts.")
				return nil
			}

			if hostname != "" {
				if _, ok := cfg.Get(hostname); !ok {
					return fmt.Errorf("not logged into %s", hostname)
				}
				hosts = []string{hostname}
			}

			sort.Strings(hosts)
			for _, h := range hosts {
				hostCfg, _ := cfg.Get(h)
				printHostStatus(f, h, hostCfg)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// printHostStatus writes the human-readable status block for a single host.
func printHostStatus(f *factory.Factory, host string, hostCfg config.HostConfig) {
	fmt.Fprintf(f.IOStreams.Out, "%s: Logged in as %s (%s)\n",
		host, hostCfg.User, hostCfg.GitProtocol)

	keyringStatus := "no"
	if _, err := f.Keyring.Get("bitbottle", hostCfg.User); err == nil {
		keyringStatus = "yes"
	}
	fmt.Fprintf(f.IOStreams.Out, "  Token in keyring: %s\n", keyringStatus)
}
