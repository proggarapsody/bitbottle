// Package secrets is the `admin secrets` subcommand tree.
package secrets

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/secrets/rotate"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdSecrets builds the `admin secrets` command tree.
func NewCmdSecrets(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage cluster secrets",
	}
	cmd.AddCommand(rotate.NewCmdRotate(f, nil))
	return cmd
}
