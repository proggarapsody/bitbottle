package root

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/completion"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	mcpcmd "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/tag"
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

	cmd.AddCommand(completion.NewCmdCompletion(f))
	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(repo.NewCmdRepo(f))
	cmd.AddCommand(pr.NewCmdPR(f))
	cmd.AddCommand(branch.NewCmdBranch(f))
	cmd.AddCommand(pipeline.NewCmdPipeline(f))
	cmd.AddCommand(tag.NewCmdTag(f))
	cmd.AddCommand(commit.NewCmdCommit(f))
	cmd.AddCommand(mcpcmd.NewCmdMCP(f))

	SetHelpFunc(cmd)

	return cmd
}
