package root

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/alias"
	"github.com/proggarapsody/bitbottle/pkg/cmd/api"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/codeinsights"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/completion"
	configcmd "github.com/proggarapsody/bitbottle/pkg/cmd/config"
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/context"     // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/deployment"  // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/environment" // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/extension"   // self-registers via init()
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/issue"
	mcpcmd "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	cmdProfile "github.com/proggarapsody/bitbottle/pkg/cmd/profile"
	"github.com/proggarapsody/bitbottle/pkg/cmd/project"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	searchcmd "github.com/proggarapsody/bitbottle/pkg/cmd/search"
	"github.com/proggarapsody/bitbottle/pkg/cmd/skill"
	"github.com/proggarapsody/bitbottle/pkg/cmd/tag"
	cmdVariable "github.com/proggarapsody/bitbottle/pkg/cmd/variable"
	"github.com/proggarapsody/bitbottle/pkg/cmd/webhook"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
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

			jsonMode, _ := c.Flags().GetBool("json")
			yamlMode, _ := c.Flags().GetBool("yaml")
			jqExpr, _ := c.Flags().GetString("jq")
			tmpl, _ := c.Flags().GetString("template")

			// Mutual-exclusion: pick exactly one output format.
			if jsonMode && yamlMode {
				return fmt.Errorf("--json and --yaml are mutually exclusive")
			}
			if jsonMode && tmpl != "" {
				return fmt.Errorf("--json and --template are mutually exclusive")
			}
			if yamlMode && tmpl != "" {
				return fmt.Errorf("--yaml and --template are mutually exclusive")
			}
			if jqExpr != "" && !jsonMode {
				return fmt.Errorf("--jq requires --json")
			}

			// Structured output: disable color so consumers get raw values.
			if jsonMode || yamlMode || tmpl != "" {
				f.IOStreams.SetColorEnabled(false)
			}
			return nil
		},
	}

	cmd.PersistentFlags().String("hostname", "", "Bitbucket hostname (overrides git remote)")
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().Bool("yaml", false, "Output as YAML")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
	cmd.PersistentFlags().String("template", "", "Format output with a Go template")
	cmdutil.RegisterNoColorFlag(cmd)

	cmd.AddCommand(completion.NewCmdCompletion(f))
	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(repo.NewCmdRepo(f))
	cmd.AddCommand(pr.NewCmdPR(f))
	cmd.AddCommand(branch.NewCmdBranch(f))
	cmd.AddCommand(codeinsights.NewCmdCodeInsights(f))
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
	cmd.AddCommand(searchcmd.NewCmdSearch(f))
	cmd.AddCommand(skill.NewCmdSkill(f))
	cmd.AddCommand(cmdVariable.NewCmdVariable(f))
	cmd.AddCommand(cmdProfile.NewCmdProfile(f))
	cmd.AddCommand(NewCmdStatus(f))
	cmd.AddCommand(NewCmdBrowse(f))

	// Self-registered commands (legacy fixed list above; new commands use
	// cmdregistry instead of editing this file).
	for _, sub := range cmdregistry.All(f) {
		cmd.AddCommand(sub)
	}

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
