package logs

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// Options holds parsed flags for `pipeline logs`.
type Options struct {
	Hostname string

	// Args[0] = PROJECT/REPO, Args[1] = PIPELINE-UUID, Args[2] = STEP-UUID
	Args []string
}

// NewCmdLogs builds the `pipeline logs` cobra command.
func NewCmdLogs(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "logs [PROJECT/REPO] PIPELINE-UUID STEP-UUID",
		Short: "Stream a pipeline step's log to stdout",
		Args:  cobra.RangeArgs(2, 3),
		// Build logs are routinely thousands of lines; route through
		// $PAGER on a TTY so users get scroll/search affordances.
		Annotations: map[string]string{cmdutil.PagerAnnotation: "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return logsRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func logsRun(f *factory.Factory, opts *Options) error {
	repoArgs, rest := repoarg.SplitLeadingRepo(opts.Args, 2)
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
	rc, err := pc.GetPipelineStepLog(ref.Project, ref.Slug, rest[0], rest[1])
	if err != nil {
		return err
	}
	defer rc.Close() //nolint:errcheck
	if _, err := io.Copy(f.IOStreams.Out, rc); err != nil {
		return fmt.Errorf("stream log: %w", err)
	}
	return nil
}
