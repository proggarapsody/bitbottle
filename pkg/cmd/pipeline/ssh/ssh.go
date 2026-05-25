// Package ssh implements `bitbottle pipeline ssh` subcommand group.
package ssh

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdKeyPair "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/ssh/keypair"
	cmdKnownHosts "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/ssh/knownhosts"
)

// NewCmdPipelineSSH builds the `pipeline ssh` parent command.
func NewCmdPipelineSSH(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Manage pipeline SSH configuration (Cloud only)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(cmdKeyPair.NewCmdKeyPair(f))
	cmd.AddCommand(cmdKnownHosts.NewCmdKnownHosts(f))
	return cmd
}
