package root

import (
	"github.com/spf13/cobra"

	_ "github.com/proggarapsody/bitbottle/pkg/cmd/admin" // self-registers via init()
	"github.com/proggarapsody/bitbottle/pkg/cmd/alias"
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/api"          // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/auth"         // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/branch"       // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/branchmodel"  // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/branchrule"   // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/codeinsights" // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/commit"       // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/completion"   // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/config"       // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/context"      // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/deploykey"    // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/deployment"   // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/diff"         // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/environment"  // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/extension"    // self-registers via init()
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/group"     // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/host"      // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/issue"     // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"       // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/milestone" // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/mirror"    // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline"  // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/pr"        // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/profile"   // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/project"   // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/repo"      // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/runner"    // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/search"    // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/skill"     // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/snippet"   // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/sshkey"    // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/tag"       // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/user"      // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/variable"  // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/version"   // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/webhook"   // self-registers via init()
	_ "github.com/proggarapsody/bitbottle/pkg/cmd/workspace" // self-registers via init()
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

			if skip, _ := c.Flags().GetBool("skip-tls-verify"); skip {
				f.SkipTLSOverride = true
			}
			if dbg, _ := c.Flags().GetBool("debug"); dbg {
				f.DebugHTTP = true
			}

			// Output-format contract (mutual-exclusion + --jq requires --json +
			// color disable for structured modes). Shared with the repo-override
			// PersistentPreRunE so command groups that install their own hook
			// don't silently skip these checks. See cmdutil.ValidateOutputFlags.
			return cmdutil.ValidateOutputFlags(c, f.IOStreams)
		},
	}

	cmd.PersistentFlags().String("hostname", "", "Bitbucket hostname (overrides git remote)")
	cmd.PersistentFlags().String("json", "", "Output as JSON (optionally select fields: --json field1,field2)")
	cmd.PersistentFlags().Lookup("json").NoOptDefVal = "*"
	cmd.PersistentFlags().Bool("yaml", false, "Output as YAML")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
	cmd.PersistentFlags().String("template", "", "Format output with a Go template")
	cmd.PersistentFlags().BoolP("skip-tls-verify", "k", false, "Skip TLS certificate verification for this invocation (self-signed CAs)")
	cmd.PersistentFlags().Bool("debug", false, "Log HTTP request/response details to stderr")
	cmdutil.RegisterNoColorFlag(cmd)

	cmd.AddCommand(NewCmdStatus(f))
	cmd.AddCommand(NewCmdBrowse(f))

	// Self-registered commands — packages call cmdregistry.Register from init()
	// so new commands never need to touch this file.
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
