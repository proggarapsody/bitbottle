package commit

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCommit(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Manage commits",
	}
	cmd.AddCommand(NewCmdCommitLog(f))
	cmd.AddCommand(NewCmdCommitView(f))
	return cmd
}
