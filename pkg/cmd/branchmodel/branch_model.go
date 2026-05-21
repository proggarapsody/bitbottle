// Package branchmodel implements the `bitbottle branch-model` command group.
// Branching model management is Cloud-only; Server/DC is not supported.
package branchmodel

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdBranchModel)
}

// NewCmdBranchModel builds the root `branch-model` command.
func NewCmdBranchModel(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch-model",
		Short: "Manage branching model",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted,
the repository is inferred from the "origin" git remote in the current directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdGet(f))
	cmd.AddCommand(NewCmdSet(f))
	return cmd
}
