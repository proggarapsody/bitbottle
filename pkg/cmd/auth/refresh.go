package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdAuthRefresh(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Re-validate the stored token against the live API",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, host, err := resolveAuthHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			user, err := client.GetCurrentUser()
			if err != nil {
				fmt.Fprintf(f.IOStreams.ErrOut,
					"error: token validation failed — run 'bitbottle auth login --hostname %s' to re-authenticate\n",
					host,
				)
				return err
			}

			hostCfg, _ := cfg.Get(host)
			if user.Slug != hostCfg.User {
				hostCfg.User = user.Slug
				cfg.Set(host, hostCfg)
				if err := cfg.Save(); err != nil {
					return err
				}
			}

			fmt.Fprintf(f.IOStreams.Out, "Authenticated to %s as %s\n", host, user.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
