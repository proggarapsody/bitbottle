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

			pr, err := client.CreatePR(ref.Project, ref.Slug, backend.CreatePRInput{
				Title:       title,
				Description: body,
				Draft:       draft,
				FromBranch:  currentBranch,
				ToBranch:    base,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Created pull request #%d: %s\n", pr.ID, pr.Title)
			if pr.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", pr.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Pull request title")
	cmd.Flags().StringVar(&body, "body", "", "Pull request description")
	cmd.Flags().StringVar(&base, "base", "", "Base branch")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create as draft")
	return cmd
}
