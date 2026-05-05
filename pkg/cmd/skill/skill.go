// Package skill implements the `bitbottle skill` command tree, which
// installs, refreshes, and removes the bundled bitbottle agent skill
// (the SKILL.md + references that teach AI agents how to drive
// bitbottle correctly).
//
// Why this command exists: the npm postinstall already wires the
// skill on initial install, but users who installed via Homebrew, Go
// install, deb/rpm, or the bare binary won't have it. They also need
// a discoverable way to refresh after a release rather than
// memorizing the npx incantation.
//
// The implementation is a thin, intentional wrapper over the
// vercel-labs/skills CLI. We don't reimplement its runtime detection
// — there are 50+ agent runtimes with non-trivial install paths each,
// and that ecosystem is moving fast. Shelling out keeps us aligned.
package skill

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

const (
	repoRef         = "proggarapsody/bitbottle"
	manualHintLine1 = "Install Node.js (>= 18), then run:"
	manualHintLine2 = "  npx -y skills add proggarapsody/bitbottle --global -y"
)

// NewCmdSkill returns the `skill` parent command.
func NewCmdSkill(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage the bundled AI agent skill",
		Long: `Install, refresh, or remove the bitbottle agent skill across detected
AI runtimes (Claude Code, Cursor, Codex, Gemini CLI, and ~50 others).

The skill is fetched from the GitHub repository so it always
reflects the latest published guidance. To pin it to a specific
release, install that release's binary and run ` + "`bitbottle skill install`" + `.`,
	}
	cmd.AddCommand(newCmdInstall(f))
	cmd.AddCommand(newCmdRemove(f))
	cmd.AddCommand(newCmdPath(f))
	return cmd
}

// newCmdInstall does a remove-then-add so reinstalls actually refresh
// existing content. The skills CLI's `add` is idempotent and would
// otherwise no-op if the skill was already installed at an older
// version — see the postinstall in packages/mcp-npm/install.js for
// the same rationale.
func newCmdInstall(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install or refresh the bitbottle skill (uses npx)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := exec.LookPath("npx"); err != nil {
				return missingNpxError()
			}
			// Best-effort remove first. Failure is fine — most likely
			// the skill simply wasn't installed yet.
			_ = runSkills(f, "remove", "bitbottle", "-g", "-y")
			if err := runSkills(f, "add", repoRef, "--global", "-y"); err != nil {
				return fmt.Errorf("skills add failed: %w", err)
			}
			fmt.Fprintln(f.IOStreams.Out, "bitbottle skill installed/refreshed.")
			return nil
		},
	}
}

func newCmdRemove(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove the bitbottle skill from all agent runtimes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := exec.LookPath("npx"); err != nil {
				return missingNpxError()
			}
			if err := runSkills(f, "remove", "bitbottle", "-g", "-y"); err != nil {
				return fmt.Errorf("skills remove failed: %w", err)
			}
			fmt.Fprintln(f.IOStreams.Out, "bitbottle skill removed.")
			return nil
		},
	}
}

// newCmdPath prints the canonical install root. Useful when an agent
// reports it can't find the skill — users can confirm the file
// landed where their runtime expects it.
func newCmdPath(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the canonical install root for the bitbottle skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(f.IOStreams.Out, "~/.agents/skills/bitbottle")
			fmt.Fprintln(f.IOStreams.Out, "  (symlinked into per-runtime locations like ~/.claude/skills/bitbottle)")
			return nil
		},
	}
}

// runSkills shells out to `npx -y skills <args...>`. Output is
// streamed to the user's TTY so they see install progress (and any
// errors from the skills CLI itself) without us mediating.
//
// Uses CommandContext with context.Background — we don't impose a
// timeout because skill installation can legitimately take a while
// on first run (npm pulls the skills package into its cache).
func runSkills(f *factory.Factory, args ...string) error {
	full := append([]string{"-y", "skills"}, args...)
	c := exec.CommandContext(context.Background(), "npx", full...)
	c.Stdout = f.IOStreams.Out
	c.Stderr = f.IOStreams.ErrOut
	return c.Run()
}

func missingNpxError() error {
	return fmt.Errorf(
		"npx not found on PATH; install Node.js (>= 18) and re-run.\n%s\n%s",
		manualHintLine1, manualHintLine2,
	)
}
