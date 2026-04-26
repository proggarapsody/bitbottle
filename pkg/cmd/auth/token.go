package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdAuthToken(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print the stored PAT for the resolved host",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, host, err := resolveAuthHostname(f, hostname)
			if err != nil {
				return err
			}

			hostCfg, _ := cfg.Get(host)
			if hostCfg.OAuthToken == "" {
				return fmt.Errorf("no token stored for %s", host)
			}

			fmt.Fprintln(f.IOStreams.Out, hostCfg.OAuthToken)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
