// Package perms is the root of the `perms` subcommand tree.
// Permission management is a Bitbucket Server / Data Center feature only —
// invocations against Cloud surfaces a typed ErrUnsupportedOnHost via the
// backend.AsPermissionsClient accessor.
package perms

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/project"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/repo"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdPerms)
}

// NewCmdPerms builds the `perms` command tree.
func NewCmdPerms(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perms",
		Short: "Manage permissions (Bitbucket Server / DC only)",
		Long: `Manage project and repository permissions on Bitbucket Server / Data Center.

Bitbucket Cloud uses a different permissions model and is not yet supported;
calling these subcommands against Cloud returns a typed "unsupported on host" error.`,
	}
	cmd.AddCommand(project.NewCmdProject(f))
	cmd.AddCommand(repo.NewCmdRepo(f))
	return cmd
}
