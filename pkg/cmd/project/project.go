// Package project implements the `bitbottle project` command group.
// The group contains:
//   - `list WORKSPACE` — Cloud-only, lists projects inside a workspace.
//   - `server-list` / `view` / `create` / `edit` / `delete` — Server/DC only,
//     manage the project namespace on Bitbucket Server / Data Center.
package project

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/project/list"
)

func NewCmdProject(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Bitbucket projects",
	}
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdView(f))
	cmd.AddCommand(NewCmdEdit(f))
	cmd.AddCommand(NewCmdDelete(f))
	cmd.AddCommand(NewCmdServerList(f))
	return cmd
}
