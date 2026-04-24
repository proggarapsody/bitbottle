package auth

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdAuthStatus(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	_ = hostname
	return cmd
}
