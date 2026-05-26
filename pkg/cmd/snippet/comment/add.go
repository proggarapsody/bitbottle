package comment

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	snippetlist "github.com/proggarapsody/bitbottle/pkg/cmd/snippet/list"
)

// AddOptions carries parsed flags for `snippet comment add`.
type AddOptions struct {
	Hostname  string
	Workspace string
	SnippetID string
	Body      string
}

// NewCmdAdd constructs the cobra command. The runF override lets tests inject
// their own runner without standing up a real backend.
func NewCmdAdd(f *factory.Factory, runF ...func(*AddOptions) error) *cobra.Command {
	opts := &AddOptions{}
	cmd := &cobra.Command{
		Use:   "add SNIPPET_ID [WORKSPACE]",
		Short: "Add a comment to a snippet",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SnippetID = args[0]
			if len(args) > 1 {
				opts.Workspace = args[1]
			}
			if len(runF) > 0 && runF[0] != nil {
				return runF[0](opts)
			}
			return addRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().StringVar(&opts.Body, "body", "", "Comment body text (required)")
	_ = cmd.MarkFlagRequired("body")
	return cmd
}

func addRun(f *factory.Factory, opts *AddOptions) error {
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
	c, err := sc.AddSnippetComment(workspace, opts.SnippetID, opts.Body)
	if err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Comment %d added to snippet %s\n", c.ID, opts.SnippetID)
	return nil
}
