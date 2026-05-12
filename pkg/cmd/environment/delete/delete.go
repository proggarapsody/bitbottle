package delete

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `environment delete`.
type Options struct {
	Hostname string
	Confirm  bool

	// Args[0] = PROJECT/REPO, Args[1] = ENV-UUID
	Args []string
}

// NewCmdDelete builds the `environment delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO ENV-UUID",
		Short: "Delete a deployment environment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func deleteRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}

	envUUID := opts.Args[1]

	if !opts.Confirm {
		if !f.IOStreams.IsStdoutTTY() {
			return fmt.Errorf("--confirm required when not running interactively")
		}
		fmt.Fprintf(f.IOStreams.Out, "Delete environment %q? [y/N] ", envUUID)
		var answer string
		if _, err := fmt.Fscan(f.IOStreams.In, &answer); err != nil || (answer != "y" && answer != "Y") {
			fmt.Fprintln(f.IOStreams.Out, "Cancelled.")
			return nil
		}
	}

	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	dc, err := backend.AsDeploymentClient(client, ref.Host)
	if err != nil {
		return err
	}
	if err := dc.DeleteEnvironment(ref.Project, ref.Slug, envUUID); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Deleted environment %s\n", envUUID)
	return nil
}
