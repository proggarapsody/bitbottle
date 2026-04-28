package root

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/alias"
	"github.com/proggarapsody/bitbottle/pkg/cmd/api"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/completion"
	configcmd "github.com/proggarapsody/bitbottle/pkg/cmd/config"
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
	cmd.AddCommand(api.NewCmdAPI(f))
	cmd.AddCommand(configcmd.NewCmdConfig(f))
	cmd.AddCommand(mcpcmd.NewCmdMCP(f))

	// Alias must be added last so it can see the names of every other built-in.
	cmd.AddCommand(alias.NewCmdAlias(f, builtinNames(cmd)))

	SetHelpFunc(cmd)

	return cmd
}

// builtinNames returns the set of top-level command names (and their direct
// aliases) registered on root. Used to prevent user aliases from shadowing
// built-ins.
func builtinNames(root *cobra.Command) []string {
	var names []string
	for _, c := range root.Commands() {
		names = append(names, c.Name())
		names = append(names, c.Aliases...)
	}
	return names
}
