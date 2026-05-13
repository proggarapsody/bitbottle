// Package logging is the `admin logging` subcommand tree.
package logging

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/logging/get"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/logging/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdLogging builds the `admin logging` command tree.
func NewCmdLogging(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logging",
		Short: "Manage log level and async logging settings",
	}
	cmd.AddCommand(get.NewCmdGet(f, nil))
	cmd.AddCommand(set.NewCmdSet(f, nil))
	return cmd
}
