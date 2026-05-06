// Package webhook is the root of the `webhook` subcommand tree.
package webhook

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdCreate "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/create"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/delete"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/list"
	cmdView "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/view"
)

// NewCmdWebhook builds the root `webhook` command. Subcommands live in their
// own subpackages (gh-CLI style).
func NewCmdWebhook(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Manage repository webhooks",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(cmdList.NewCmdList(f, nil))
	cmd.AddCommand(cmdView.NewCmdView(f, nil))
	cmd.AddCommand(cmdCreate.NewCmdCreate(f, nil))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f, nil))
	return cmd
}
