package mergecheck

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// SetOptions holds parsed flags for `code-insights merge-check set`.
type SetOptions struct {
	Hostname    string
	ReportKey   string
	MustPass    bool
	MinSeverity string
	// Args[0]=PROJECT/REPO  Args[1]=KEY
	Args []string
}

// NewCmdSet builds the `code-insights merge-check set` cobra command.
func NewCmdSet(f *factory.Factory, runF func(*SetOptions) error) *cobra.Command {
	opts := &SetOptions{}
	cmd := &cobra.Command{
		Use:   "set [PROJECT/REPO] KEY",
		Short: "Create or update merge-check config — EXPERIMENTAL",
		Long: `Create or update a merge-check configuration on Bitbucket Server.

EXPERIMENTAL: This command uses a partly undocumented API.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return setRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.ReportKey, "report-key", "", "Code Insights report key this check applies to (required)")
	cmd.Flags().BoolVar(&opts.MustPass, "must-pass", false, "Block merge when the report does not pass (required)")
	cmd.Flags().StringVar(&opts.MinSeverity, "min-severity", "", "Minimum annotation severity to block merge: LOW, MEDIUM, HIGH, CRITICAL")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	_ = cmd.MarkFlagRequired("report-key")
	return cmd
}

func setRun(f *factory.Factory, opts *SetOptions) error {
	repoArgs, rest := repoarg.SplitLeadingRepo(opts.Args, 1)
	ref, err := factory.ResolveTarget(f, repoArgs, opts.Hostname)
	if err != nil {
		return err
	}
	key := rest[0]
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	ci, err := backend.AsCodeInsightsClient(client, ref.Host)
	if err != nil {
		return err
	}
	in := backend.MergeCheckInput{
		Key:         key,
		ReportKey:   opts.ReportKey,
		MustPass:    opts.MustPass,
		MinSeverity: strings.ToUpper(opts.MinSeverity),
	}
	if err := ci.SetMergeCheck(ref.Project, ref.Slug, key, in); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Set merge check %q\n", key)
	return nil
}
