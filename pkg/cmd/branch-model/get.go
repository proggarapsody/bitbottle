package branchmodel

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdGet builds the `branch-model get` cobra command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	var hostname string
	var jsonFlag bool
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
			if jsonFlag || cmd.Flags().Changed("json") {
				return json.NewEncoder(f.IOStreams.Out).Encode(model)
			}
			return printBranchModel(f, model)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
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
