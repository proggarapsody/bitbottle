package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoCreate(f *factory.Factory) *cobra.Command {
	var project, description string
	var private bool
	var jsonFields string
	var jqExpr string

	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return fmt.Errorf("--project is required")
			}

			name := args[0]

			host, err := resolveHostname(f, "")
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			r, err := client.CreateRepo(project, backend.CreateRepoInput{
				Name:        name,
				SCM:         "git",
				Public:      !private,
				Description: description,
			})
			if err != nil {
				return err
			}

			if jsonFields != "" || jqExpr != "" {
				p := repoFields(f, jsonFields, jqExpr)
				p.SetSingleItem()
				p.AddItem(r)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Created repository %s/%s\n", r.Namespace, r.Slug)
			if r.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", r.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "Project key")
	cmd.Flags().StringVar(&description, "description", "", "Repository description")
	cmd.Flags().BoolVar(&private, "private", true, "Make repository private")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	return cmd
}
