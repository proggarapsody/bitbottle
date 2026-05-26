package comment

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	snippetlist "github.com/proggarapsody/bitbottle/pkg/cmd/snippet/list"
)

// DeleteOptions carries parsed flags for `snippet comment delete`.
type DeleteOptions struct {
	Hostname  string
	Workspace string
	SnippetID string
	CommentID int
	Confirm   bool
}

// NewCmdDelete constructs the cobra command. The runF override lets tests inject
// their own runner without standing up a real backend.
func NewCmdDelete(f *factory.Factory, runF ...func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete SNIPPET_ID COMMENT_ID [WORKSPACE]",
		Short: "Delete a comment from a snippet",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.SnippetID = args[0]
			id, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid comment ID %q: must be an integer", args[1])
			}
			opts.CommentID = id
			if len(args) > 2 {
				opts.Workspace = args[2]
			}
			if len(runF) > 0 && runF[0] != nil {
				return runF[0](opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip interactive confirmation")
	return cmd
}

func deleteRun(f *factory.Factory, opts *DeleteOptions) error {
	// Non-TTY guard: require --confirm when stdout is not a TTY.
	if !f.IOStreams.IsStdoutTTY() && !opts.Confirm {
		return fmt.Errorf("--confirm required when not running interactively")
	}
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
	if err := sc.DeleteSnippetComment(workspace, opts.SnippetID, opts.CommentID); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Comment %d deleted from snippet %s\n", opts.CommentID, opts.SnippetID)
	return nil
}
