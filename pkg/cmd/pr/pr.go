package pr

import (
	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdPR(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests",
	}
	cmd.AddCommand(NewCmdPRList(f))
	cmd.AddCommand(NewCmdPRView(f))
	cmd.AddCommand(NewCmdPRCreate(f))
	cmd.AddCommand(NewCmdPRMerge(f))
	cmd.AddCommand(NewCmdPRApprove(f))
	cmd.AddCommand(NewCmdPRDiff(f))
	cmd.AddCommand(NewCmdPRCheckout(f))
	return cmd
}
