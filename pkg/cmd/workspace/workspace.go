// Package workspace implements the `bitbottle workspace` command group.
// Workspaces are a Bitbucket Cloud concept; the optional WorkspaceClient
// interface gates these commands so an invocation against a Server/DC host
// surfaces a typed ErrUnsupportedOnHost rather than a runtime panic.
package workspace

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdAudit "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/audit"
	cmdhook "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/hook"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/list"
	cmdMemberList "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/member"
)

func NewCmdWorkspace(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "List Bitbucket Cloud workspaces (Cloud only)",
	}
	cmd.AddCommand(cmdList.NewCmdList(f, nil))

	memberCmd := &cobra.Command{
		Use:   "member",
		Short: "Manage workspace members (Cloud only)",
	}
	memberCmd.AddCommand(cmdMemberList.NewCmdList(f, nil))
	cmd.AddCommand(memberCmd)

	hookCmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage workspace webhooks (Cloud only)",
	}
	hookCmd.AddCommand(cmdhook.NewCmdList(f, nil))
	hookCmd.AddCommand(cmdhook.NewCmdCreate(f, nil))
	hookCmd.AddCommand(cmdhook.NewCmdDelete(f, nil))
	cmd.AddCommand(hookCmd)

	cmd.AddCommand(cmdAudit.NewCmdAudit(f, nil))

	return cmd
}
