package repo

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	cmdClone "github.com/proggarapsody/bitbottle/pkg/cmd/repo/clone"
	cmdCreate "github.com/proggarapsody/bitbottle/pkg/cmd/repo/create"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/repo/delete"
	cmdDownload "github.com/proggarapsody/bitbottle/pkg/cmd/repo/download"
	cmdEdit "github.com/proggarapsody/bitbottle/pkg/cmd/repo/edit"
	cmdFile "github.com/proggarapsody/bitbottle/pkg/cmd/repo/file"
	cmdFork "github.com/proggarapsody/bitbottle/pkg/cmd/repo/fork"
	cmdHook "github.com/proggarapsody/bitbottle/pkg/cmd/repo/hook"
	cmdLabel "github.com/proggarapsody/bitbottle/pkg/cmd/repo/label"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/repo/list"
	cmdPRSettings "github.com/proggarapsody/bitbottle/pkg/cmd/repo/pr-settings"
	cmdRename "github.com/proggarapsody/bitbottle/pkg/cmd/repo/rename"
	cmdSetDefault "github.com/proggarapsody/bitbottle/pkg/cmd/repo/set-default"
	cmdSetDefaultBranch "github.com/proggarapsody/bitbottle/pkg/cmd/repo/set-default-branch"
	cmdSync "github.com/proggarapsody/bitbottle/pkg/cmd/repo/sync"
	cmdTransfer "github.com/proggarapsody/bitbottle/pkg/cmd/repo/transfer"
	cmdTree "github.com/proggarapsody/bitbottle/pkg/cmd/repo/tree"
	cmdView "github.com/proggarapsody/bitbottle/pkg/cmd/repo/view"
	cmdVisibility "github.com/proggarapsody/bitbottle/pkg/cmd/repo/visibility"
	cmdWatcher "github.com/proggarapsody/bitbottle/pkg/cmd/repo/watcher"
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
	cmd.AddCommand(cmdEdit.NewCmdEdit(f))
	cmd.AddCommand(cmdList.NewCmdList(f))
	cmd.AddCommand(cmdView.NewCmdView(f))
	cmd.AddCommand(cmdCreate.NewCmdCreate(f))
	cmd.AddCommand(cmdDelete.NewCmdDelete(f))
	cmd.AddCommand(cmdClone.NewCmdClone(f))
	cmd.AddCommand(cmdSetDefault.NewCmdSetDefault(f))
	cmd.AddCommand(cmdRename.NewCmdRename(f))
	forkCmd := &cobra.Command{
		Use:   "fork",
		Short: "Create and list repository forks",
	}
	forkCmd.AddCommand(cmdFork.NewCmdForkCreate(f))
	forkCmd.AddCommand(cmdFork.NewCmdForkList(f))
	cmd.AddCommand(forkCmd)
	cmd.AddCommand(cmdTransfer.NewCmdTransfer(f))
	cmd.AddCommand(cmdFile.NewCmdFile(f))
	cmd.AddCommand(cmdTree.NewCmdTree(f))
	cmd.AddCommand(cmdLabel.NewCmdLabel(f))
	cmd.AddCommand(cmdWatcher.NewCmdWatcher(f))
	cmd.AddCommand(cmdVisibility.NewCmdVisibility(f))
	cmd.AddCommand(cmdSetDefaultBranch.NewCmdSetDefaultBranch(f))
	cmd.AddCommand(cmdPRSettings.NewCmdPRSettings(f))
	cmd.AddCommand(cmdDownload.NewCmdDownload(f))
	cmd.AddCommand(cmdHook.NewCmdHook(f))
	cmd.AddCommand(cmdSync.NewCmdSync(f))
	return cmd
}
