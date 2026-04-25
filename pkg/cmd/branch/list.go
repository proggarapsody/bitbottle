package branch

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdBranchList(f *factory.Factory) *cobra.Command {
	var limit int
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List branches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := f.ResolveRef(args[0], hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			branches, err := client.ListBranches(ref.Project, ref.Slug, limit)
			if err != nil {
				return err
			}

			p := branchFields(f, jsonFields, jqExpr)
			for _, b := range branches {
				p.AddItem(b)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of branches")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func branchFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Branch] {
	p := format.New[backend.Branch](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Branch]{Name: "name", Header: "NAME", Extract: func(b backend.Branch) any { return b.Name }})
	p.AddField(format.Field[backend.Branch]{Name: "default", Header: "DEFAULT", Extract: func(b backend.Branch) any { return b.IsDefault }})
	p.AddField(format.Field[backend.Branch]{Name: "hash", Header: "HASH", Extract: func(b backend.Branch) any {
		if len(b.LatestHash) > 8 {
			return b.LatestHash[:8]
		}
		return b.LatestHash
	}})
	return p
}
