// Package watch implements `bb pipeline watch UUID`.
package watch

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// terminalStates are pipeline states that signal completion.
// The Cloud adapter flattens "COMPLETED" into its result name
// ("SUCCESSFUL", "FAILED", "STOPPED", "EXPIRED", "ERROR"), so
// we match on those result names, not on "COMPLETED".
var terminalStates = map[string]bool{
	"SUCCESSFUL": true,
	"FAILED":     true,
	"STOPPED":    true,
	"EXPIRED":    true,
	"ERROR":      true,
}

// Options holds parsed flags for `pipeline watch`.
type Options struct {
	Hostname string
	Interval int

	// Args[0] = PROJECT/REPO, Args[1] = UUID
	Args []string
}

// NewCmdWatch builds the `pipeline watch` cobra command.
func NewCmdWatch(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "watch PROJECT/REPO UUID",
		Short: "Watch a pipeline until it completes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return watchRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Interval, "interval", 5, "Polling interval in seconds")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func watchRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
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

	uuid := opts.Args[1]
	out := f.IOStreams.Out
	lastState := ""

	for {
		pl, err := pc.GetPipeline(ref.Project, ref.Slug, uuid)
		if err != nil {
			return err
		}

		if pl.State != lastState {
			fmt.Fprintf(out, "Pipeline #%d state: %s\n", pl.BuildNumber, pl.State)
			lastState = pl.State
		}

		if terminalStates[pl.State] {
			if pl.State == "SUCCESSFUL" {
				return nil
			}
			return fmt.Errorf("pipeline ended with state: %s", pl.State)
		}

		time.Sleep(time.Duration(opts.Interval) * time.Second)
	}
}
