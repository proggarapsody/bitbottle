package pr

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRDefaultReviewerList builds the `pr default-reviewer list` cobra command.
func NewCmdPRDefaultReviewerList(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List default reviewers for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			dr, err := backend.AsDefaultReviewerClient(client, ref.Host)
			if err != nil {
				return err
			}
			reviewers, err := dr.ListDefaultReviewers(ref.Project, ref.Slug)
			if err != nil {
				return err
			}
			p := defaultReviewerFields(f, format.ConfigFromCmd(cmd))
			for _, r := range reviewers {
				p.AddItem(r)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func defaultReviewerFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.DefaultReviewer] {
	p := format.New[backend.DefaultReviewer](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.DefaultReviewer]{
		Name:    "userSlug",
		Header:  "USER",
		Extract: func(r backend.DefaultReviewer) any { return r.UserSlug },
	})
	p.AddField(format.Field[backend.DefaultReviewer]{
		Name:    "displayName",
		Header:  "DISPLAY NAME",
		Extract: func(r backend.DefaultReviewer) any { return r.DisplayName },
	})
	p.AddField(format.Field[backend.DefaultReviewer]{
		Name:    "emailAddress",
		Header:  "EMAIL",
		Extract: func(r backend.DefaultReviewer) any { return r.EmailAddress },
	})
	return p
}
