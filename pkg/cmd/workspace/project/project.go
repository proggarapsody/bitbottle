// Package project implements the `bitbottle workspace project` command group.
// Workspace projects are a Bitbucket Cloud concept gated by the
// CloudProjectClient optional interface.
package project

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdWorkspaceProject(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage workspace projects (Cloud)",
	}
	cmd.AddCommand(NewCmdCreate(f))
	cmd.AddCommand(NewCmdView(f))
	cmd.AddCommand(NewCmdEdit(f))
	cmd.AddCommand(NewCmdDelete(f))
	return cmd
}
