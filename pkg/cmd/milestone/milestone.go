// Package milestone implements the `bitbottle milestone` command group.
// Milestones are a Bitbucket Cloud issue-tracker feature gated by the
// MilestoneClient optional interface; invocations against Server/DC surface
// a typed ErrUnsupportedOnHost.
package milestone

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdMilestone)
}

func NewCmdMilestone(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestone",
		Short: "List and view issue milestones (Cloud)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdMilestoneList(f))
	cmd.AddCommand(NewCmdMilestoneView(f))
	return cmd
}
