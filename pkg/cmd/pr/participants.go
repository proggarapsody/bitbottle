package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// NewCmdPRParticipants returns the "pr participants" sub-command.
func NewCmdPRParticipants(f *factory.Factory) *cobra.Command {
	var hostnameFlag string

	cmd := &cobra.Command{
		Use:         "participants PR_ID",
		Short:       "List participants in a pull request",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{cmdutil.PagerAnnotation: "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			pp, err := backend.AsPRParticipantClient(client, ref.Host)
			if err != nil {
				return err
			}

			participants, err := pp.ListPRParticipants(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			p := prParticipantsFields(f, format.ConfigFromCmd(cmd))
			for _, participant := range participants {
				p.AddItem(participant)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func prParticipantsFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.PRParticipant] {
	isTTY := f.IOStreams.IsStdoutTTY()
	p := format.New[backend.PRParticipant](f.IOStreams.Out, isTTY, cfg)

	p.AddField(format.Field[backend.PRParticipant]{
		Name:   "role",
		Header: "ROLE",
		Extract: func(e backend.PRParticipant) any {
			return e.Role
		},
	})

	p.AddField(format.Field[backend.PRParticipant]{
		Name:   "display_name",
		Header: "DISPLAY_NAME",
		Extract: func(e backend.PRParticipant) any {
			return e.User.DisplayName
		},
	})

	p.AddField(format.Field[backend.PRParticipant]{
		Name:   "username",
		Header: "USERNAME",
		Extract: func(e backend.PRParticipant) any {
			return e.User.Slug
		},
	})

	p.AddField(format.Field[backend.PRParticipant]{
		Name:   "approved",
		Header: "APPROVED",
		Extract: func(e backend.PRParticipant) any {
			return fmt.Sprintf("%v", e.Approved)
		},
	})

	return p
}
