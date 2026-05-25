// Package activity implements the `bitbottle issue activity` subcommand.
// Issue activity is a Bitbucket Cloud-only feature gated by the
// IssueActivityClient optional interface; Server/DC returns ErrUnsupportedOnHost.
package activity

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// NewCmdIssueActivity returns the Cobra command for `issue activity`.
func NewCmdIssueActivity(f *factory.Factory) *cobra.Command {
	var hostname string
	var limit int
	cmd := &cobra.Command{
		Use:   "activity ISSUE_ID [PROJECT/REPO]",
		Short: "List the activity/change history of a Cloud issue",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(limit); err != nil {
				return err
			}

			// Discriminate: first arg is always the issue ID, optional second is repo.
			idArg := args[0]
			var repoArgs []string
			if len(args) == 2 {
				repoArgs = []string{args[1]}
			}

			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}

			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			ac, err := backend.AsIssueActivityClient(client, ref.Host)
			if err != nil {
				return err
			}

			changes, err := ac.ListIssueActivity(ref.Project, ref.Slug, id, limit)
			if err != nil {
				return err
			}

			p := issueActivityFields(f, format.ConfigFromCmd(cmd))
			for _, c := range changes {
				p.AddItem(c)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "Maximum number of activity entries")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func issueActivityFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.IssueChange] {
	p := format.New[backend.IssueChange](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.IssueChange]{Name: "id", Header: "ID", Extract: func(c backend.IssueChange) any {
		return fmt.Sprintf("#%d", c.ID)
	}})
	p.AddField(format.Field[backend.IssueChange]{Name: "kind", Header: "KIND", Extract: func(c backend.IssueChange) any {
		return c.Kind
	}})
	p.AddField(format.Field[backend.IssueChange]{Name: "old_val", Header: "OLD", Extract: func(c backend.IssueChange) any {
		return c.OldVal
	}})
	p.AddField(format.Field[backend.IssueChange]{Name: "new_val", Header: "NEW", Extract: func(c backend.IssueChange) any {
		return c.NewVal
	}})
	p.AddField(format.Field[backend.IssueChange]{Name: "created_on", Header: "DATE", Extract: func(c backend.IssueChange) any {
		return c.CreatedOn.Format("2006-01-02")
	}})
	p.AddField(format.Field[backend.IssueChange]{Name: "user", Header: "USER", Extract: func(c backend.IssueChange) any {
		if c.User.DisplayName != "" {
			return c.User.DisplayName
		}
		return c.User.Slug
	}})
	return p
}
