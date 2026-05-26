package defaultreviewer

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// RemoveOptions carries parsed flags for `workspace project default-reviewer remove`.
type RemoveOptions struct {
	Hostname   string
	Workspace  string
	ProjectKey string
	AccountID  string
	Confirm    bool
}

// NewCmdRemove constructs the `workspace project default-reviewer remove` cobra command.
func NewCmdRemove(f *factory.Factory, runF func(*RemoveOptions) error) *cobra.Command {
	opts := &RemoveOptions{}
	cmd := &cobra.Command{
		Use:   "remove WORKSPACE PROJECT_KEY",
		Short: "Remove a default reviewer from a workspace project (Cloud only)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Workspace = args[0]
			opts.ProjectKey = args[1]
			if runF != nil {
				return runF(opts)
			}
			return removeRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.AccountID, "user", "", "Account ID of the user to remove (required)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	_ = cmd.MarkFlagRequired("user")
	return cmd
}

func removeRun(f *factory.Factory, opts *RemoveOptions) error {
	if !opts.Confirm && !f.IOStreams.IsStdoutTTY() {
		return fmt.Errorf("--confirm required when not running interactively")
	}
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
	if err := rc.RemoveProjectDefaultReviewer(opts.Workspace, opts.ProjectKey, opts.AccountID); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Removed %s from default reviewers on project %s in %s.\n",
		opts.AccountID, opts.ProjectKey, opts.Workspace)
	return nil
}
