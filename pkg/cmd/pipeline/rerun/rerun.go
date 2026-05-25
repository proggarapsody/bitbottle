// Package rerun implements `bitbottle pipeline rerun UUID [PROJECT/REPO]`.
package rerun

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `pipeline rerun`.
type Options struct {
	Hostname string

	// Args[0] = source pipeline UUID; Args[1] = optional PROJECT/REPO
	Args []string
}

// NewCmdRerun builds the `pipeline rerun` cobra command.
func NewCmdRerun(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "rerun UUID [PROJECT/REPO]",
		Short: "Re-run a finished pipeline at the same commit",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runRerun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func runRerun(f *factory.Factory, opts *Options) error {
	// UUID is always args[0]; optional PROJECT/REPO is args[1].
	uuid := opts.Args[0]
	repoArgs := opts.Args[1:]

	ref, err := factory.ResolveTarget(f, repoArgs, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPipelineClient(client, ref.Host)
	if err != nil {
		return err
	}

	pl, err := pc.RerunPipeline(ref.Project, ref.Slug, uuid)
	if err != nil {
		return err
	}

	out := f.IOStreams.Out
	if f.IOStreams.IsStdoutTTY() {
		fmt.Fprintf(out, "Pipeline #%d queued: %s\n", pl.BuildNumber, pl.WebURL)
	} else {
		fmt.Fprintf(out, "%d\t%s\n", pl.BuildNumber, pl.WebURL)
	}
	return nil
}
