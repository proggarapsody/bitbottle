package tag

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdTag(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdTagList(f))
	cmd.AddCommand(NewCmdTagCreate(f))
	cmd.AddCommand(NewCmdTagDelete(f))
	return cmd
}
