package set

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `environment variable set`.
type Options struct {
	Hostname string
	Secured  bool

	// Args[0] = PROJECT/REPO, Args[1] = ENV-UUID, Args[2] = KEY, Args[3] = VALUE
	Args []string
}

// NewCmdSet builds the `environment variable set` cobra command.
func NewCmdSet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "set PROJECT/REPO ENV-UUID KEY VALUE",
		Short: "Set an environment variable",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return setRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Secured, "secured", false, "Mark as secured (value redacted on read)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func setRun(f *factory.Factory, opts *Options) error {
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
	v, err := dc.SetEnvVariable(ref.Project, ref.Slug, opts.Args[1], backend.EnvVariableInput{
		Key:     opts.Args[2],
		Value:   opts.Args[3],
		Secured: opts.Secured,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Set variable %q (UUID: %s)\n", v.Key, v.UUID)
	return nil
}
