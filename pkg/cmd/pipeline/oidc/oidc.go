// Package oidc implements `bitbottle pipeline oidc` subcommand group.
package oidc

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPipelineOIDC builds the `pipeline oidc` parent command.
func NewCmdPipelineOIDC(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oidc",
		Short: "View pipeline OIDC configuration (Cloud only)",
	}
	cmd.AddCommand(NewCmdConfig(f, nil))
	cmd.AddCommand(NewCmdKeys(f, nil))
	return cmd
}
