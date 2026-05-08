package annotation

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DeleteOptions holds parsed flags for `code-insights annotation delete`.
type DeleteOptions struct {
	Hostname string
	// Args[0]=PROJECT/REPO  Args[1]=HASH  Args[2]=KEY
	Args []string
}

// NewCmdDelete builds the `code-insights annotation delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO HASH KEY",
		Short: "Delete all annotations under a Code Insights report",
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

func deleteRun(f *factory.Factory, opts *DeleteOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args[:1], opts.Hostname)
	if err != nil {
		return err
	}
	hash := opts.Args[1]
	key := opts.Args[2]
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	ci, err := backend.AsCodeInsightsClient(client, ref.Host)
	if err != nil {
		return err
	}
	if err := ci.DeleteAnnotations(ref.Project, ref.Slug, hash, key); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Deleted annotations for report %q\n", key)
	return nil
}
