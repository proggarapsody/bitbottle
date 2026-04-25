package pr

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRCreate(f *factory.Factory) *cobra.Command {
	var title, body, base string
	var draft bool
	var jsonFields string
	var jqExpr string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := resolveRepoRef(f, args, "")
			if err != nil {
				return err
			}
			ref.Project = strings.ToUpper(ref.Project)

			g := git.New(f.GitRunner())
			currentBranch, err := g.CurrentBranch()
			if err != nil {
				return err
			}

			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if base == "" {
				return fmt.Errorf("--base is required")
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			p, err := client.CreatePR(ref.Project, ref.Slug, backend.CreatePRInput{
				Title:       title,
				Description: body,
				Draft:       draft,
				FromBranch:  currentBranch,
				ToBranch:    base,
			})
			if err != nil {
				return err
			}

			if jsonFields != "" || jqExpr != "" {
				printer := prFields(f, jsonFields, jqExpr)
				printer.SetSingleItem()
				printer.AddItem(p)
				return printer.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Created pull request #%d: %s\n", p.ID, p.Title)
			if p.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", p.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Pull request title")
	cmd.Flags().StringVar(&body, "body", "", "Pull request description")
	cmd.Flags().StringVar(&base, "base", "", "Base branch")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create as draft")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	return cmd
}
