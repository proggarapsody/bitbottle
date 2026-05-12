package deployment

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/list"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/view"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdDeployment)
}

// NewCmdDeployment builds the root `deployment` command.
func NewCmdDeployment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deployment",
		Short: "Manage Bitbucket Cloud deployments",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(list.NewCmdList(f, nil))
	cmd.AddCommand(view.NewCmdView(f, nil))
	return cmd
}
