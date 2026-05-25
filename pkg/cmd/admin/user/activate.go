package user

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ActivateOptions holds parsed flags for `admin user activate`.
type ActivateOptions struct {
	Hostname string
	Slug     string
}

// NewCmdUserActivate builds the `admin user activate` cobra command.
func NewCmdUserActivate(f *factory.Factory, runF func(*ActivateOptions) error) *cobra.Command {
	opts := &ActivateOptions{}
	cmd := &cobra.Command{
		Use:   "activate SLUG",
		Short: "Activate a Bitbucket Server user account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Slug = args[0]
			if runF != nil {
				return runF(opts)
			}
			return activateRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func activateRun(f *factory.Factory, opts *ActivateOptions) error {
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
	if err := ac.ActivateUser(opts.Slug); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "User %s activated.\n", opts.Slug)
	return nil
}
