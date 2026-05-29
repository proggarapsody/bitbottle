// Package host implements the `bitbottle host` command group.
package host

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/host/info"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdHost)
}

func NewCmdHost(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host",
		Short: "Inspect connected Bitbucket hosts",
	}
	cmd.AddCommand(info.NewCmdHostInfo(f))
	return cmd
}
