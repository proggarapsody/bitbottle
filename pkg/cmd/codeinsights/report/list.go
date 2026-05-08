package report

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ListOptions holds parsed flags for `code-insights report list`.
type ListOptions struct {
	Hostname   string
	JSONFields string
	JQExpr     string
	// Args[0]=PROJECT/REPO  Args[1]=HASH
	Args []string
}

// NewCmdList builds the `code-insights report list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO HASH",
		Short: "List Code Insights reports for a commit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.JSONFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&opts.JQExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func listRun(f *factory.Factory, opts *ListOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args[:1], opts.Hostname)
	if err != nil {
		return err
	}
	hash := opts.Args[1]
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	ci, err := backend.AsCodeInsightsClient(client, ref.Host)
	if err != nil {
		return err
	}
	reports, err := ci.ListReports(ref.Project, ref.Slug, hash)
	if err != nil {
		return err
	}
	p := reportFields(f, opts.JSONFields, opts.JQExpr)
	for _, r := range reports {
		p.AddItem(r)
	}
	return p.Render()
}
