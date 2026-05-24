package pat

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdRevoke builds `auth pat revoke TOKEN_ID [--hostname H] [--confirm]`.
func NewCmdRevoke(f *factory.Factory) *cobra.Command {
	var hostname string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "revoke TOKEN_ID",
		Short: "Revoke a personal access token on Bitbucket Server/DC",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tokenID := args[0]

			if !confirm {
				if !f.IOStreams.IsStdoutTTY() {
					return fmt.Errorf("--confirm is required in non-interactive mode")
				}
				fmt.Fprintf(f.IOStreams.Out, "Revoke PAT %s? [y/N] ", tokenID)
				scanner := bufio.NewScanner(f.IOStreams.In)
				var answer string
				if scanner.Scan() {
					answer = strings.TrimSpace(scanner.Text())
				}
				if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
					fmt.Fprintln(f.IOStreams.Out, "Revoke aborted.")
					return nil
				}
			}

			host, userSlug, err := resolveHostAndUser(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			pc, err := backend.AsPATClient(client, host)
			if err != nil {
				return err
			}

			if err := pc.RevokePAT(userSlug, tokenID); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Revoked PAT %s\n", tokenID)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	return cmd
}
