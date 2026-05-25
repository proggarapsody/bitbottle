package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// RenameOptions holds parsed flags for `admin user rename`.
type RenameOptions struct {
	Hostname string
	OldSlug  string
	NewSlug  string
}

// NewCmdUserRename builds the `admin user rename` cobra command.
func NewCmdUserRename(f *factory.Factory, runF func(*RenameOptions) error) *cobra.Command {
	opts := &RenameOptions{}
	cmd := &cobra.Command{
		Use:   "rename OLD_SLUG NEW_SLUG",
		Short: "Rename a Bitbucket Server user",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.OldSlug = args[0]
			opts.NewSlug = args[1]
			if runF != nil {
				return runF(opts)
			}
			return renameRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func renameRun(f *factory.Factory, opts *RenameOptions) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	ac, err := backend.AsAdminClient(client, host)
	if err != nil {
		return err
	}
	if err := ac.RenameUser(opts.OldSlug, opts.NewSlug); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "User %s renamed to %s.\n", opts.OldSlug, opts.NewSlug)
	return nil
}
