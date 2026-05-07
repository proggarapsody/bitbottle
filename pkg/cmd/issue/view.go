package issue

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdIssueView(f *factory.Factory) *cobra.Command {
	var jsonFields, jqExpr, hostname string
	cmd := &cobra.Command{
		Use:   "view [PROJECT/REPO] ID",
		Short: "View a single issue",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			ic, err := backend.AsIssueClient(client, ref.Host)
			if err != nil {
				return err
			}
			issue, err := ic.GetIssue(ref.Project, ref.Slug, id)
			if err != nil {
				return err
			}
			p := issueViewFields(f, jsonFields, jqExpr)
			p.SetSingleItem()
			p.AddItem(issue)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// splitIDArg returns (repoArgs, idArg). Forms accepted:
//
//	[PROJECT/REPO ID]  — explicit repo + id
//	[ID]               — id only; repo inferred from BaseRepo
//
// We discriminate by arg count rather than by parsing the last token,
// matching `pr view` / `pr diff` ergonomics so the issue group stays
// consistent with the rest of the CLI.
func splitIDArg(args []string) ([]string, string) {
	if len(args) == 1 {
		return nil, args[0]
	}
	return []string{args[0]}, args[1]
}
