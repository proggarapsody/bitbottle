package repo

import (
	"fmt"

	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdRepoCreate(f *factory.Factory) *cobra.Command {
	var project, description string
	var private bool

	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "Project key")
	cmd.Flags().StringVar(&description, "description", "", "Repository description")
	cmd.Flags().BoolVar(&private, "private", true, "Make repository private")
	_ = project
	_ = description
	_ = private
	return cmd
}
