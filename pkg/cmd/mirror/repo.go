package mirror

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdMirrorRepo returns the `mirror repo` sub-group command.
func NewCmdMirrorRepo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage mirrored repos (Server/DC)",
	}
	cmd.AddCommand(NewCmdMirrorRepoList(f, nil))
	return cmd
}
