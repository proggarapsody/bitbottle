package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `environment variable delete`.
type Options struct {
	Hostname string

	// Args[0] = PROJECT/REPO, Args[1] = ENV-UUID, Args[2] = VAR-UUID
	Args []string
}

// NewCmdDelete builds the `environment variable delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO ENV-UUID VAR-UUID",
		Short: "Delete an environment variable",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func deleteRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	dc, err := backend.AsDeploymentClient(client, ref.Host)
	if err != nil {
		return err
	}
	envUUID := opts.Args[1]
	varUUID := opts.Args[2]
	if err := dc.DeleteEnvVariable(ref.Project, ref.Slug, envUUID, varUUID); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Deleted variable %s\n", varUUID)
	return nil
}
