package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRReopen builds the `pr reopen PR_ID` command. Reopen is the reverse
// of `pr decline` — it returns a previously declined PR to the OPEN state.
//
// Bitbucket Server / Data Center exposes a dedicated /reopen endpoint.
// Bitbucket Cloud has no reopen primitive (Atlassian BCLOUD-23807) and so
// returns a typed ErrUnsupportedOnHost via backend.AsPRReopener.
func NewCmdPRReopen(f *factory.Factory) *cobra.Command {
	var hostnameFlag string

	cmd := &cobra.Command{
		Use:   "reopen PR_ID",
		Short: "Reopen a declined pull request (Bitbucket Server / DC only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			reopener, err := backend.AsPRReopener(client, ref.Host)
			if err != nil {
				return err
			}

			if err := reopener.ReopenPR(ref.Project, ref.Slug, prID); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Reopened pull request #%d\n", prID)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	return cmd
}
