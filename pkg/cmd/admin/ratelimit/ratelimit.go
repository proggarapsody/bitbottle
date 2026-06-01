// Package ratelimit is the `admin rate-limit` subcommand tree.
package ratelimit

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/ratelimit/get"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/ratelimit/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdRateLimit builds the `admin rate-limit` command tree.
func NewCmdRateLimit(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-limit",
		Short: "Manage Server rate-limit configuration",
	}
	cmd.AddCommand(get.NewCmdGet(f, nil))
	cmd.AddCommand(set.NewCmdSet(f, nil))
	return cmd
}
