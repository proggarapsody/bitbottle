package report

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// ViewOptions holds parsed flags for `code-insights report view`.
type ViewOptions struct {
	Output   format.OutputConfig
	Hostname string
	// Args[0]=PROJECT/REPO  Args[1]=HASH  Args[2]=KEY
	Args []string
}

// NewCmdView builds the `code-insights report view` cobra command.
func NewCmdView(f *factory.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{}
	cmd := &cobra.Command{
		Use:   "view [PROJECT/REPO] HASH KEY",
		Short: "View a single Code Insights report",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func viewRun(f *factory.Factory, opts *ViewOptions) error {
	repoArgs, rest := repoarg.SplitLeadingRepo(opts.Args, 2)
	ref, err := factory.ResolveTarget(f, repoArgs, opts.Hostname)
	if err != nil {
		return err
	}
	hash := rest[0]
	key := rest[1]
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	ci, err := resolveCIAdapter(client, ref.Host)
	if err != nil {
		return err
	}
	r, err := ci.GetReport(ref.Project, ref.Slug, hash, key)
	if err != nil {
		return err
	}
	p := reportFields(f, opts.Output)
	p.SetSingleItem()
	p.AddItem(r)
	return p.Render()
}
