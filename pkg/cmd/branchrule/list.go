package branchrule

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// NewCmdList builds the `branch-rule list` cobra command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List branch restriction rules for a repository",
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
			rules, listErr := br.ListBranchRules(ref.Project, ref.Slug)
			if listErr != nil && len(rules) == 0 {
				return listErr
			}
			p := branchRuleFields(f, format.ConfigFromCmd(cmd))
			for _, r := range rules {
				p.AddItem(r)
			}
			if err := p.Render(); err != nil {
				return err
			}
			cmdutil.PartialWarn(f.IOStreams.ErrOut, len(rules), listErr)
			return listErr
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func branchRuleFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.BranchRule] {
	p := format.New[backend.BranchRule](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.BranchRule]{Name: "id", Header: "ID", Extract: func(r backend.BranchRule) any { return r.ID }})
	p.AddField(format.Field[backend.BranchRule]{Name: "kind", Header: "KIND", Extract: func(r backend.BranchRule) any { return r.Kind }})
	p.AddField(format.Field[backend.BranchRule]{Name: "pattern", Header: "PATTERN", Extract: func(r backend.BranchRule) any { return r.Pattern }})
	p.AddField(format.Field[backend.BranchRule]{Name: "value", Header: "VALUE", Extract: func(r backend.BranchRule) any { return r.Value }})
	return p
}
