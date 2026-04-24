package root

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
)

func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "bitbottle",
		Short:         "Bitbucket CLI",
		Long:          "bitbottle is a CLI for self-hosted Bitbucket Server/Data Center.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.PersistentFlags().String("hostname", "", "Bitbucket hostname (overrides git remote)")

	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(repo.NewCmdRepo(f))
	cmd.AddCommand(pr.NewCmdPR(f))

	return cmd
}
