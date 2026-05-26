package defaultreviewer

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ListOptions carries parsed flags for `workspace project default-reviewer list`.
type ListOptions struct {
	Output     format.OutputConfig
	Hostname   string
	Workspace  string
	ProjectKey string
	Limit      int
}

// NewCmdList constructs the `workspace project default-reviewer list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list WORKSPACE PROJECT_KEY",
		Short: "List default reviewers for a workspace project (Cloud only)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			opts.Workspace = args[0]
			opts.ProjectKey = args[1]
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of results (0 = no cap)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	rc, err := backend.AsWorkspaceProjectDefaultReviewerClient(client, host)
	if err != nil {
		return err
	}
	reviewers, err := rc.ListProjectDefaultReviewers(opts.Workspace, opts.ProjectKey, opts.Limit)
	if err != nil {
		return err
	}

	p := projectDefaultReviewerFields(f, opts.Output)
	for _, r := range reviewers {
		p.AddItem(r)
	}
	return p.Render()
}

func projectDefaultReviewerFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.ProjectDefaultReviewer] {
	p := format.New[backend.ProjectDefaultReviewer](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.ProjectDefaultReviewer]{
		Name:    "account_id",
		Header:  "ACCOUNT ID",
		Extract: func(r backend.ProjectDefaultReviewer) any { return r.AccountID },
	})
	p.AddField(format.Field[backend.ProjectDefaultReviewer]{
		Name:    "display_name",
		Header:  "DISPLAY NAME",
		Extract: func(r backend.ProjectDefaultReviewer) any { return r.DisplayName },
	})
	p.AddField(format.Field[backend.ProjectDefaultReviewer]{
		Name:    "nickname",
		Header:  "NICKNAME",
		Extract: func(r backend.ProjectDefaultReviewer) any { return r.Nickname },
	})
	return p
}
