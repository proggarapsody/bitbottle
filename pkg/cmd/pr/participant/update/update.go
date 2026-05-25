package update

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	prshared "github.com/proggarapsody/bitbottle/pkg/cmd/pr/shared"
)

// NewCmdPRParticipantUpdate returns the "pr participant update" sub-command.
func NewCmdPRParticipantUpdate(f *factory.Factory) *cobra.Command {
	var (
		hostnameFlag   string
		userFlag       string
		approve        bool
		unapprove      bool
		requestChanges bool
	)

	cmd := &cobra.Command{
		Use:   "update PR_ID --user ACCOUNT_ID (--approve|--unapprove|--request-changes) [PROJECT/REPO]",
		Short: "Update a pull request participant's approval state",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			prID, err := prshared.ParsePRID(args[0])
			if err != nil {
				return err
			}

			if !approve && !unapprove && !requestChanges {
				return fmt.Errorf("one of --approve, --unapprove, --request-changes is required")
			}

			var state string
			switch {
			case approve:
				state = "approved"
			case requestChanges:
				state = "changes_requested"
			case unapprove:
				state = ""
			}

			var repoArg []string
			if len(args) > 1 {
				repoArg = args[1:]
			}
			ref, err := factory.ResolveTarget(f, repoArg, hostnameFlag)
			if err != nil {
				return err
			}
			ref.Project = strings.ToUpper(ref.Project)

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			updater, err := backend.AsPRParticipantUpdater(client, ref.Host)
			if err != nil {
				return err
			}

			p, err := updater.UpdatePRParticipant(ref.Project, ref.Slug, prID, userFlag, state)
			if err != nil {
				return err
			}

			displayState := p.State
			if displayState == "" {
				displayState = "unapproved"
			}
			fmt.Fprintf(f.IOStreams.Out, "Participant %s state set to %s\n", p.User.DisplayName, displayState)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&userFlag, "user", "", "Account ID of the participant to update")
	_ = cmd.MarkFlagRequired("user")
	cmd.Flags().BoolVar(&approve, "approve", false, "Set state to approved")
	cmd.Flags().BoolVar(&unapprove, "unapprove", false, "Set state to unapproved (neutral)")
	cmd.Flags().BoolVar(&requestChanges, "request-changes", false, "Set state to changes_requested")
	cmd.MarkFlagsMutuallyExclusive("approve", "unapprove", "request-changes")

	return cmd
}
