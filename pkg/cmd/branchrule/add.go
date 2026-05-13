package branchrule

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdAdd builds the `branch-rule add` cobra command.
func NewCmdAdd(f *factory.Factory) *cobra.Command {
	var kind, pattern, hostname string
	var value int
	cmd := &cobra.Command{
		Use:   "add [PROJECT/REPO]",
		Short: "Add a branch restriction rule to a repository",
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
			br, err := backend.AsBranchRuleClient(client, ref.Host)
			if err != nil {
				return err
			}
			added, err := br.AddBranchRule(ref.Project, ref.Slug, backend.BranchRuleInput{
				Kind:    kind,
				Pattern: pattern,
				Value:   value,
			})
			if err != nil {
				return err
			}
			p := branchRuleFields(f, format.ConfigFromCmd(cmd))
			p.AddItem(added)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "", "Branch restriction kind (required)")
	cmd.Flags().StringVar(&pattern, "pattern", "", "Branch pattern to restrict (required)")
	cmd.Flags().IntVar(&value, "value", 0, "Numeric value for the rule (e.g. required approvals)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	_ = cmd.MarkFlagRequired("kind")
	_ = cmd.MarkFlagRequired("pattern")
	return cmd
}
