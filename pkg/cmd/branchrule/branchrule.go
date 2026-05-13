// Package branchrule implements the `bitbottle branch-rule` command group.
// Branch restriction rules are Cloud-only; Server/DC is not supported.
package branchrule

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdBranchRule)
}

// NewCmdBranchRule builds the root `branch-rule` command.
func NewCmdBranchRule(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch-rule",
		Short: "Manage Cloud branch restriction rules",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted,
the repository is inferred from the "origin" git remote in the current directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdAdd(f))
	cmd.AddCommand(NewCmdDelete(f))
	return cmd
}
