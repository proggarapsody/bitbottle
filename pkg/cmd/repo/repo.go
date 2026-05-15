package repo

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Manage repositories",
		Annotations: map[string]string{
			"help:arguments": `A repository is supplied as PROJECT/REPO. Use --hostname to
disambiguate when multiple Bitbucket hosts are configured.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdRepoList(f))
	cmd.AddCommand(NewCmdRepoView(f))
	cmd.AddCommand(NewCmdRepoCreate(f))
	cmd.AddCommand(NewCmdRepoDelete(f))
	cmd.AddCommand(NewCmdRepoClone(f))
	cmd.AddCommand(NewCmdRepoSetDefault(f))
	cmd.AddCommand(NewCmdRepoRename(f))
	forkCmd := &cobra.Command{
		Use:   "fork",
		Short: "Create and list repository forks",
	}
	forkCmd.AddCommand(NewCmdRepoForkCreate(f))
	forkCmd.AddCommand(NewCmdRepoForkList(f))
	cmd.AddCommand(forkCmd)
	cmd.AddCommand(NewCmdRepoTransfer(f))
	cmd.AddCommand(NewCmdRepoFile(f))
	cmd.AddCommand(NewCmdRepoTree(f))
	cmd.AddCommand(NewCmdRepoWatcher(f))
	return cmd
}
