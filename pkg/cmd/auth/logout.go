package auth

import (
	"fmt"

	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func NewCmdAuthLogout(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out of a Bitbucket host",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented")
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	_ = hostname
	return cmd
}
