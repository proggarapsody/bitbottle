// Package banner implements `admin banner get`, `admin banner set`, and
// `admin banner clear`.
package banner

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/banner/clear"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/banner/get"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/banner/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdBanner builds the `admin banner` command group.
func NewCmdBanner(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "banner",
		Short: "Manage the site-wide announcement banner (Server/DC only)",
	}
	cmd.AddCommand(get.NewCmdBannerGet(f, nil))
	cmd.AddCommand(set.NewCmdBannerSet(f, nil))
	cmd.AddCommand(clear.NewCmdBannerClear(f, nil))
	return cmd
}
