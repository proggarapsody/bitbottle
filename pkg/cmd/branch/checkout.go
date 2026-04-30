package branch

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdBranchCheckout(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "checkout NAME",
		Short: "Check out a branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = hostname
			name := args[0]

			g := git.New(f.GitRunner())

			if err := g.Fetch("origin", name); err != nil {
				return err
			}

			stdout, _, err := f.GitRunner().Run("branch", "--list", name)
			if err != nil {
				return err
			}

			if strings.TrimSpace(stdout) != "" {
				return g.Checkout(name)
			}

			_, _, err = f.GitRunner().Run("checkout", "-b", name, "--track", "origin/"+name)
			return err
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
