package protect

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// ListOptions holds parsed flags for `branch protect list`.
type ListOptions struct {
	Hostname string
	Limit    int
	Output   format.OutputConfig

	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdList builds the `branch protect list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{Limit: 30}
	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List branch restrictions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(opts.Limit); err != nil {
				return err
			}
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of restrictions to list")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	bp, err := backend.AsBranchProtector(client, ref.Host)
	if err != nil {
		return err
	}
	got, listErr := bp.ListBranchProtections(ref.Project, ref.Slug, opts.Limit)
	if listErr != nil && len(got) == 0 {
		return listErr
	}
	p := fields(f, opts.Output)
	for _, r := range got {
		p.AddItem(r)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(got), listErr)
	return listErr
}
