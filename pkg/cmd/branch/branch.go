package branch

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdBranch(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch",
		Short: "Manage branches",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdBranchList(f))
	cmd.AddCommand(NewCmdBranchDelete(f))
	cmd.AddCommand(NewCmdBranchCreate(f))
	cmd.AddCommand(NewCmdBranchCheckout(f))
	return cmd
}
