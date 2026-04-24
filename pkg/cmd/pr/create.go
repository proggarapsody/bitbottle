package pr

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPRCreate(f *factory.Factory) *cobra.Command {
	var title, body, base string
	var draft bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Pull request title")
	cmd.Flags().StringVar(&body, "body", "", "Pull request description")
	cmd.Flags().StringVar(&base, "base", "", "Base branch")
	cmd.Flags().BoolVar(&draft, "draft", false, "Create as draft")
	_ = title
	_ = body
	_ = base
	_ = draft
	return cmd
}
