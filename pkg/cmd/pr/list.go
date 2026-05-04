package pr

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRList(f *factory.Factory) *cobra.Command {
	var state string
	var limit int
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List pull requests",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			prs, err := client.ListPRs(ref.Project, ref.Slug, strings.ToUpper(state), limit)
			if err != nil {
				return err
			}

			p := prFields(f, jsonFields, jqExpr)
			for _, pr := range prs {
				p.AddItem(pr)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&state, "state", "open", "State filter: open, closed, merged")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of pull requests")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
