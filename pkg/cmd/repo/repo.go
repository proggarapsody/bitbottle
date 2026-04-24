package repo

import (
	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdRepo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories",
	}
	cmd.AddCommand(NewCmdRepoList(f))
	cmd.AddCommand(NewCmdRepoView(f))
	cmd.AddCommand(NewCmdRepoCreate(f))
	cmd.AddCommand(NewCmdRepoDelete(f))
	cmd.AddCommand(NewCmdRepoClone(f))
	return cmd
}
