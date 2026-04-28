package commit

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCommit(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Manage commits",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdCommitLog(f))
	cmd.AddCommand(NewCmdCommitView(f))
	cmd.AddCommand(NewCmdCommitStatus(f))
	return cmd
}
