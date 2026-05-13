package environment

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/create"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/delete"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/list"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable" //nolint:staticcheck
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdEnvironment)
}

// NewCmdEnvironment builds the root `environment` command.
func NewCmdEnvironment(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "environment",
		Short: "Manage Bitbucket Cloud environments",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(list.NewCmdList(f, nil))
	cmd.AddCommand(create.NewCmdCreate(f, nil))
	cmd.AddCommand(delete.NewCmdDelete(f, nil))
	cmd.AddCommand(variable.NewCmdVariable(f))
	return cmd
}
