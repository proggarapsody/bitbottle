// Package delete implements `bitbottle snippet delete`.
package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	snippetlist "github.com/proggarapsody/bitbottle/pkg/cmd/snippet/list"
)

// Options carries parsed flags for `snippet delete`.
type Options struct {
	Hostname  string
	Workspace string
	ID        string
	Confirm   bool
}

// NewCmdDelete constructs the cobra command. The runF override lets tests inject
// their own runner without standing up a real backend.
func NewCmdDelete(f *factory.Factory, runF ...func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "delete SNIPPET_ID",
		Short: "Delete a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ID = args[0]
			if len(runF) > 0 && runF[0] != nil {
				return runF[0](opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Workspace, "workspace", "", "Workspace slug (defaults to authenticated user)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip interactive confirmation")
	return cmd
}

func deleteRun(f *factory.Factory, opts *Options) error {
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
	if err := sc.DeleteSnippet(workspace, opts.ID); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Snippet %s deleted\n", opts.ID)
	return nil
}
