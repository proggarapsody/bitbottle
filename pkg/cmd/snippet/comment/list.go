package comment

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	snippetlist "github.com/proggarapsody/bitbottle/pkg/cmd/snippet/list"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions carries parsed flags for `snippet comment list`.
type ListOptions struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
	SnippetID string
	Limit     int
}

// NewCmdList constructs the cobra command. The runF override lets tests inject
// their own runner without standing up a real backend.
func NewCmdList(f *factory.Factory, runF ...func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list SNIPPET_ID [WORKSPACE]",
		Short: "List comments on a snippet",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SnippetID = args[0]
			if len(args) > 1 {
				opts.Workspace = args[1]
			}
			opts.Output = format.ConfigFromCmd(cmd)
			if err := cmdutil.ValidatePositiveLimit(opts.Limit); err != nil {
				return err
			}
			if len(runF) > 0 && runF[0] != nil {
				return runF[0](opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum number of comments (0 = no cap)")
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	workspace, err := snippetlist.ResolveWorkspace(f, host, opts.Workspace)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	sc, err := backend.AsSnippetClient(client, host)
	if err != nil {
		return err
	}
	comments, err := sc.ListSnippetComments(workspace, opts.SnippetID, opts.Limit)
	if err != nil {
		return err
	}
	p := snippetCommentListFields(f, opts.Output)
	for _, c := range comments {
		p.AddItem(c)
	}
	return p.Render()
}

func snippetCommentListFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.SnippetComment] {
	p := format.New[backend.SnippetComment](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.SnippetComment]{Name: "id", Header: "ID", Extract: func(c backend.SnippetComment) any { return c.ID }})
	p.AddField(format.Field[backend.SnippetComment]{Name: "author", Header: "AUTHOR", Extract: func(c backend.SnippetComment) any { return c.Author }})
	p.AddField(format.Field[backend.SnippetComment]{Name: "body", Header: "BODY", Extract: func(c backend.SnippetComment) any { return c.Body }})
	p.AddField(format.Field[backend.SnippetComment]{Name: "created_on", Header: "CREATED", Extract: func(c backend.SnippetComment) any { return c.CreatedOn }})
	p.AddField(format.Field[backend.SnippetComment]{Name: "url", Header: "URL", Aliases: []string{"webURL", "link"}, Extract: func(c backend.SnippetComment) any { return c.WebURL }})
	return p
}
