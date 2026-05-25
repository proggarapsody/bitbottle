// Package config implements `bitbottle pipeline config get/enable/disable`.
package config

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdConfig builds the `pipeline config` subgroup command.
func NewCmdConfig(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage pipeline configuration for a repository",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdGet(f, nil))
	cmd.AddCommand(NewCmdEnable(f, nil))
	cmd.AddCommand(NewCmdDisable(f, nil))
	return cmd
}
