package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdAuthLogout(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of a Bitbucket host",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, host, err := resolveAuthHostname(f, hostname)
			if err != nil {
				return err
			}
			hostCfg, _ := cfg.Get(host)

			cfg.Remove(host)
			if err := cfg.Save(); err != nil {
				return err
			}

			// Best-effort keyring deletion.
			if krErr := f.Keyring.Delete("bitbottle", hostCfg.User); krErr != nil {
				fmt.Fprintf(f.IOStreams.ErrOut, "warning: could not remove token from keyring: %v\n", krErr)
			}

			fmt.Fprintf(f.IOStreams.Out, "Logged out of %s\n", host)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
