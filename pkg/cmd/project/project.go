// Package project implements the `bitbottle project` command group.
// Projects here mean Bitbucket Cloud projects (logical groupings of repos
// inside a workspace) — not Bitbucket Server projects, which are the
// namespace itself and don't need a separate listing command.
package project

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/project/list"
)

func NewCmdProject(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "List Bitbucket Cloud projects within a workspace (Cloud only)",
	}
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	return cmd
}
