// Package mail implements `admin mail get` and `admin mail set`.
package mail

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/mail/get"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/mail/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdMail builds the `admin mail` command group.
func NewCmdMail(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mail",
		Short: "Manage the mail server configuration (Server/DC only)",
	}
	cmd.AddCommand(get.NewCmdMailGet(f, nil))
	cmd.AddCommand(set.NewCmdMailSet(f, nil))
	return cmd
}
