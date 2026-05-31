package mergecheck

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// GetOptions holds parsed flags for `code-insights merge-check get`.
type GetOptions struct {
	Hostname string
	// Args[0]=PROJECT/REPO  Args[1]=KEY
	Args []string
}

// NewCmdGet builds the `code-insights merge-check get` cobra command.
func NewCmdGet(f *factory.Factory, runF func(*GetOptions) error) *cobra.Command {
	opts := &GetOptions{}
	cmd := &cobra.Command{
		Use:   "get [PROJECT/REPO] KEY",
		Short: "Get the current merge-check configuration — EXPERIMENTAL",
		Long: `Get the current merge-check configuration on Bitbucket Server.

EXPERIMENTAL: This command uses a partly undocumented API.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return getRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func getRun(f *factory.Factory, opts *GetOptions) error {
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
	mc, err := ci.GetMergeCheck(ref.Project, ref.Slug, key)
	if err != nil {
		return err
	}
	// Always emit JSON for merge-check get (structured output only).
	enc := json.NewEncoder(f.IOStreams.Out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(mc); err != nil {
		return fmt.Errorf("encoding merge check: %w", err)
	}
	return nil
}
