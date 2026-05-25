// Package download implements the `bitbottle repo download` subcommand group.
// Repository downloads are a Bitbucket Cloud feature gated by the
// RepoDownloadClient optional interface; invocations against Server/DC surface
// a typed ErrUnsupportedOnHost.
package download

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdDownload(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Manage repository downloads (Cloud)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdDownloadList(f))
	cmd.AddCommand(NewCmdDownloadUpload(f))
	cmd.AddCommand(NewCmdDownloadGet(f))
	cmd.AddCommand(NewCmdDownloadDelete(f))
	return cmd
}
