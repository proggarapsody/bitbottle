// Package artifact implements the `pipeline artifact` subcommand group.
package artifact

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdArtifact builds the `pipeline artifact` parent command.
func NewCmdArtifact(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Manage pipeline artifacts",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdDownload(f, nil))
	return cmd
}
