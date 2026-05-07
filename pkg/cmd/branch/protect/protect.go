// Package protect is the root of the `branch protect` subcommand tree.
// Branch protection is a Bitbucket Server / Data Center feature only —
// invocations against Cloud surface a typed ErrUnsupportedOnHost via the
// backend.AsBranchProtector accessor.
package protect

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdProtect builds the `branch protect` command tree.
func NewCmdProtect(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "protect",
		Short: "Manage branch restrictions (Bitbucket Server / DC only)",
		Long: `Manage branch restrictions on Bitbucket Server / Data Center.
Bitbucket Cloud has a different "branch restrictions" model that is not
yet wired up; calling these subcommands against Cloud returns a typed
"unsupported on host" error.`,
	}
	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdCreate(f, nil))
	cmd.AddCommand(NewCmdDelete(f, nil))
	return cmd
}
