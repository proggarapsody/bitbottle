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
	cmdIPAllowlist "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/ipallowlist"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/list"
	cmdMemberList "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/member"
	cmdPerms "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/perms"
	cmdPipelineVar "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/pipelinevar"
	cmdProject "github.com/proggarapsody/bitbottle/pkg/cmd/workspace/project"
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

	pipelineVarCmd := &cobra.Command{
		Use:   "pipeline-variable",
		Short: "Manage workspace pipeline variables (Cloud only)",
	}
	pipelineVarCmd.AddCommand(cmdPipelineVar.NewCmdList(f, nil))
	pipelineVarCmd.AddCommand(cmdPipelineVar.NewCmdGet(f, nil))
	pipelineVarCmd.AddCommand(cmdPipelineVar.NewCmdSet(f, nil))
	pipelineVarCmd.AddCommand(cmdPipelineVar.NewCmdDelete(f, nil))
	cmd.AddCommand(pipelineVarCmd)

	cmd.AddCommand(cmdAudit.NewCmdAudit(f, nil))
	cmd.AddCommand(cmdProject.NewCmdWorkspaceProject(f))
	cmd.AddCommand(cmdPerms.NewCmdWorkspacePerms(f))

	ipallowlistCmd := &cobra.Command{
		Use:   "ipallowlist",
		Short: "Manage workspace IP allowlists (Cloud only)",
	}
	ipallowlistCmd.AddCommand(cmdIPAllowlist.NewCmdList(f, nil))
	ipallowlistCmd.AddCommand(cmdIPAllowlist.NewCmdAdd(f, nil))
	ipallowlistCmd.AddCommand(cmdIPAllowlist.NewCmdDelete(f, nil))
	cmd.AddCommand(ipallowlistCmd)

	return cmd
}
