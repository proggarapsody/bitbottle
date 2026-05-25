package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DeactivateOptions holds parsed flags for `admin user deactivate`.
type DeactivateOptions struct {
	Hostname string
	Slug     string
}

// NewCmdUserDeactivate builds the `admin user deactivate` cobra command.
func NewCmdUserDeactivate(f *factory.Factory, runF func(*DeactivateOptions) error) *cobra.Command {
	opts := &DeactivateOptions{}
	cmd := &cobra.Command{
		Use:   "deactivate SLUG",
		Short: "Deactivate a Bitbucket Server user account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Slug = args[0]
			if runF != nil {
				return runF(opts)
			}
			return deactivateRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func deactivateRun(f *factory.Factory, opts *DeactivateOptions) error {
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
	if err := ac.DeactivateUser(opts.Slug); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "User %s deactivated.\n", opts.Slug)
	return nil
}
