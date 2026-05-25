package issue

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdIssueAttachment returns the `issue attachment` subcommand group.
func NewCmdIssueAttachment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Manage attachments on a Bitbucket Cloud issue",
	}
	cmd.AddCommand(NewCmdIssueAttachmentList(f))
	cmd.AddCommand(NewCmdIssueAttachmentDelete(f))
	return cmd
}

func issueAttachmentFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.IssueAttachment] {
	p := format.New[backend.IssueAttachment](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.IssueAttachment]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(a backend.IssueAttachment) any { return a.Name },
	})
	p.AddField(format.Field[backend.IssueAttachment]{
		Name:    "size",
		Header:  "SIZE",
		Extract: func(a backend.IssueAttachment) any { return a.Size },
	})
	p.AddField(format.Field[backend.IssueAttachment]{
		Name:    "url",
		Header:  "URL",
		Extract: func(a backend.IssueAttachment) any { return a.Links.Self },
	})
	return p
}

// NewCmdIssueAttachmentList builds `issue attachment list [PROJECT/REPO] ISSUE_ID`.
func NewCmdIssueAttachmentList(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO] ISSUE_ID",
		Short: "List attachments on an issue",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
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
			ia, err := backend.AsIssueAttacher(client, ref.Host)
			if err != nil {
				return err
			}
			attachments, err := ia.ListIssueAttachments(ref.Project, ref.Slug, id)
			if err != nil {
				return err
			}
			p := issueAttachmentFields(f, format.ConfigFromCmd(cmd))
			for _, a := range attachments {
				p.AddItem(a)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// splitAttachmentArgs handles: [PROJECT/REPO] ISSUE_ID FILENAME
// Returns (repoArgs, issueID, filename).
func splitAttachmentArgs(args []string) ([]string, string, string) {
	if len(args) == 3 {
		return []string{args[0]}, args[1], args[2]
	}
	return nil, args[0], args[1]
}

// NewCmdIssueAttachmentDelete builds `issue attachment delete [PROJECT/REPO] ISSUE_ID FILENAME`.
func NewCmdIssueAttachmentDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO] ISSUE_ID FILENAME",
		Short: "Delete an attachment from an issue",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg, filename := splitAttachmentArgs(args)
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
			ia, err := backend.AsIssueAttacher(client, ref.Host)
			if err != nil {
				return err
			}
			if err := ia.DeleteIssueAttachment(ref.Project, ref.Slug, id, filename); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Deleted attachment %q from issue #%d\n", filename, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
