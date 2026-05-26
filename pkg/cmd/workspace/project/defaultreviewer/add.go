package defaultreviewer

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// AddOptions carries parsed flags for `workspace project default-reviewer add`.
type AddOptions struct {
	Hostname   string
	Workspace  string
	ProjectKey string
	AccountID  string
}

// NewCmdAdd constructs the `workspace project default-reviewer add` cobra command.
func NewCmdAdd(f *factory.Factory, runF func(*AddOptions) error) *cobra.Command {
	opts := &AddOptions{}
	cmd := &cobra.Command{
		Use:   "add WORKSPACE PROJECT_KEY",
		Short: "Add a default reviewer to a workspace project (Cloud only)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = args[0]
			opts.ProjectKey = args[1]
			if runF != nil {
				return runF(opts)
			}
			return addRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.AccountID, "user", "", "Account ID of the user to add as default reviewer (required)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	_ = cmd.MarkFlagRequired("user")
	return cmd
}

func addRun(f *factory.Factory, opts *AddOptions) error {
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
	if err := rc.AddProjectDefaultReviewer(opts.Workspace, opts.ProjectKey, opts.AccountID); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Added %s as default reviewer on project %s in %s.\n",
		opts.AccountID, opts.ProjectKey, opts.Workspace)
	return nil
}
