package branchmodel

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdGet builds the `branch-model get` cobra command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "get [PROJECT/REPO]",
		Short: "Show branching model for a repository",
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
			bm, err := backend.AsBranchModelClient(client, ref.Host)
			if err != nil {
				return err
			}
			model, err := bm.GetBranchModel(ref.Project, ref.Slug)
			if err != nil {
				return err
			}
			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := format.New[backend.BranchModel](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
				p.SetSingleItem()
				p.AddField(format.Field[backend.BranchModel]{
					Name:    "development",
					Header:  "DEVELOPMENT",
					Extract: func(m backend.BranchModel) any { return m.Development.Name },
				})
				p.AddField(format.Field[backend.BranchModel]{
					Name:   "production",
					Header: "PRODUCTION",
					Extract: func(m backend.BranchModel) any {
						if m.Production != nil {
							return m.Production.Name
						}
						return nil
					},
				})
				p.AddField(format.Field[backend.BranchModel]{
					Name:     "branch_types",
					Header:   "BRANCH_TYPES",
					JSONOnly: true,
					Extract:  func(m backend.BranchModel) any { return m.BranchTypes },
				})
				p.AddItem(model)
				return p.Render()
			}
			return printBranchModel(f, model)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func printBranchModel(f *factory.Factory, m backend.BranchModel) error {
	out := f.IOStreams.Out
	dev := m.Development
	fmt.Fprintf(out, "Development: %s", dev.Name)
	if dev.UseMainbranch {
		fmt.Fprintf(out, " (main branch)")
	}
	fmt.Fprintln(out)
	if m.Production != nil {
		prod := m.Production
		fmt.Fprintf(out, "Production:  %s", prod.Name)
		if prod.UseMainbranch {
			fmt.Fprintf(out, " (main branch)")
		}
		fmt.Fprintln(out)
	}
	if len(m.BranchTypes) > 0 {
		fmt.Fprintln(out, "Branch types:")
		for _, bt := range m.BranchTypes {
			fmt.Fprintf(out, "  %-12s %s\n", bt.Kind, bt.Prefix)
		}
	}
	return nil
}
