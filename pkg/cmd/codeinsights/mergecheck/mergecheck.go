// Package mergecheck is the `code-insights merge-check` subcommand group.
// The merge-check API is partly undocumented in Bitbucket Server — these
// commands are marked experimental.
package mergecheck

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdMergeCheck builds the `code-insights merge-check` command tree.
func NewCmdMergeCheck(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge-check",
		Short: "Manage Code Insights merge checks — EXPERIMENTAL",
		Long: `Manage merge-check configurations for Code Insights on Bitbucket Server.

EXPERIMENTAL: The merge-check API is partly undocumented and may change in
future versions of Bitbucket Server / Data Center. Use with caution.`,
	}
	cmd.AddCommand(NewCmdSet(f, nil))
	cmd.AddCommand(NewCmdGet(f, nil))
	cmd.AddCommand(NewCmdDelete(f, nil))
	return cmd
}
