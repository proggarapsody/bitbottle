// Package admin is the root of the `admin` subcommand tree.
// Admin operations are Bitbucket Server / Data Center features only —
// invocations against Cloud surface a typed ErrUnsupportedOnHost via the
// backend.AsAdminClient accessor.
package admin

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/cluster"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/license"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/logging"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/secrets"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/user"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdAdmin)
}

// NewCmdAdmin builds the `admin` command tree.
func NewCmdAdmin(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Administer a Bitbucket Server instance",
		Long: `Administer a Bitbucket Server / Data Center instance.

All subcommands require SYS_ADMIN permission. Standard admin tokens do not
include it; these commands must be performed by a system administrator.`,
	}
	cmd.AddCommand(secrets.NewCmdSecrets(f))
	cmd.AddCommand(logging.NewCmdLogging(f))
	cmd.AddCommand(user.NewCmdUser(f))
	cmd.AddCommand(license.NewCmdLicense(f, nil))
	cmd.AddCommand(cluster.NewCmdCluster(f, nil))
	return cmd
}
