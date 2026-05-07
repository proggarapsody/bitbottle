package root

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmdutil"

	"github.com/proggarapsody/bitbottle/pkg/cmd/alias"
	"github.com/proggarapsody/bitbottle/pkg/cmd/api"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/completion"
	configcmd "github.com/proggarapsody/bitbottle/pkg/cmd/config"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/issue"
	mcpcmd "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/pkg/cmd/project"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/skill"
	"github.com/proggarapsody/bitbottle/pkg/cmd/tag"
	"github.com/proggarapsody/bitbottle/pkg/cmd/webhook"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace"
)

func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "bitbottle",
		Short:         "Bitbucket CLI",
		Long:          "bitbottle is a CLI for self-hosted Bitbucket Server/Data Center.",
		SilenceErrors: true,
		SilenceUsage:  true,
		// PersistentPreRunE runs before every subcommand's RunE. It is the
		// single point that translates global flags into IOStreams state, so
		// --no-color works on every command without each one re-reading it.
		// Note: cobra skips this hook when --help short-circuits, so the
		// applied state is only visible during real command execution.
		PersistentPreRunE: func(c *cobra.Command, _ []string) error {
			cmdutil.ApplyNoColorFlag(c, f.IOStreams)
			return nil
		},
	}

	cmd.PersistentFlags().String("hostname", "", "Bitbucket hostname (overrides git remote)")
	cmdutil.RegisterNoColorFlag(cmd)

	cmd.AddCommand(completion.NewCmdCompletion(f))
	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(repo.NewCmdRepo(f))
	cmd.AddCommand(pr.NewCmdPR(f))
	cmd.AddCommand(branch.NewCmdBranch(f))
	cmd.AddCommand(pipeline.NewCmdPipeline(f))
	cmd.AddCommand(tag.NewCmdTag(f))
	cmd.AddCommand(webhook.NewCmdWebhook(f))
	cmd.AddCommand(commit.NewCmdCommit(f))
	cmd.AddCommand(issue.NewCmdIssue(f))
	cmd.AddCommand(api.NewCmdAPI(f))
	cmd.AddCommand(workspace.NewCmdWorkspace(f))
	cmd.AddCommand(project.NewCmdProject(f))
	cmd.AddCommand(configcmd.NewCmdConfig(f))
	cmd.AddCommand(mcpcmd.NewCmdMCP(f))
	cmd.AddCommand(skill.NewCmdSkill(f))

	cmd.AddCommand(alias.NewCmdAlias(f, builtinNames(cmd)))

	SetHelpFunc(cmd)

	// Wrap pager-annotated commands once after tree assembly so that
	// individual commands no longer need StartPager/StopPager boilerplate.
	cmdutil.EnablePagerForAnnotated(cmd, f.IOStreams)

	return cmd
}

func builtinNames(root *cobra.Command) []string {
	var names []string
	for _, c := range root.Commands() {
		names = append(names, c.Name())
		names = append(names, c.Aliases...)
	}
	return names
}
