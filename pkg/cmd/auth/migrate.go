package auth

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdAuthMigrate creates the `bitbottle auth migrate` subcommand.
// It moves any token found in hosts.yml into the OS keyring, then zeroes it
// from the config file so that subsequent saves will not write it back.
func NewCmdAuthMigrate(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate token from config file to keyring",
		Long: `Move any plaintext token from hosts.yml into the OS keyring.

After migration, bitbottle saves and loads credentials from the keyring
only. The hosts.yml file is rewritten with the token field removed.`,
		Example: `  # Migrate tokens for all configured hosts
  bitbottle auth migrate

  # Migrate a single host
  bitbottle auth migrate --hostname git.example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			var hosts []string
			if hostname != "" {
				if _, ok := cfg.Get(hostname); !ok {
					return fmt.Errorf("not logged into %s", hostname)
				}
				hosts = []string{hostname}
			} else {
				hosts = cfg.Hosts()
				sort.Strings(hosts)
			}

			for _, host := range hosts {
				hc, _ := cfg.Get(host)
				if hc.OAuthToken == "" {
					fmt.Fprintf(f.IOStreams.Out, "✓ %s: no config-file token found\n", host)
					continue
				}

				token := hc.OAuthToken
				if err := f.Keyring.Set("bitbottle", host, token); err != nil {
					return fmt.Errorf("could not store token in keyring for %s: %w", host, err)
				}

				hc.OAuthToken = ""
				cfg.Set(host, hc)
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("could not save config for %s: %w", host, err)
				}

				fmt.Fprintf(f.IOStreams.Out, "✓ %s: token migrated to keyring\n", host)
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Migrate only this hostname")
	return cmd
}
