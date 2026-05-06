package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRCheckout(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout PR_ID",
		Short: "Check out a pull request branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, "")
			if err != nil {
				return err
			}

			pr, err := client.GetPR(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			g := git.New(f.GitRunner())
			if err := g.Fetch("origin", pr.FromBranch); err != nil {
				return fmt.Errorf("run inside a local clone of %s/%s: %w", ref.Project, ref.Slug, err)
			}
			return g.Checkout(pr.FromBranch)
		},
	}
	return cmd
}
