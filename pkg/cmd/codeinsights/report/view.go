package report

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ViewOptions holds parsed flags for `code-insights report view`.
type ViewOptions struct {
	Hostname   string
	JSONFields string
	// Args[0]=PROJECT/REPO  Args[1]=HASH  Args[2]=KEY
	Args []string
}

// NewCmdView builds the `code-insights report view` cobra command.
func NewCmdView(f *factory.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{}
	cmd := &cobra.Command{
		Use:   "view PROJECT/REPO HASH KEY",
		Short: "View a single Code Insights report",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.JSONFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func viewRun(f *factory.Factory, opts *ViewOptions) error {
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
	r, err := ci.GetReport(ref.Project, ref.Slug, hash, key)
	if err != nil {
		return err
	}
	p := reportFields(f, opts.JSONFields, "")
	p.SetSingleItem()
	p.AddItem(r)
	return p.Render()
}
