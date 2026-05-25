// Package version implements the `bitbottle version` command group.
// Issue versions are a Bitbucket Cloud issue-tracker feature gated by the
// IssueVersionClient optional interface; invocations against Server/DC surface
// a typed ErrUnsupportedOnHost.
package version

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() { cmdregistry.Register(NewCmdVersion) }

func NewCmdVersion(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Manage issue versions (Cloud)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdVersionList(f))
	cmd.AddCommand(NewCmdVersionView(f))
	cmd.AddCommand(NewCmdVersionCreate(f))
	cmd.AddCommand(NewCmdVersionDelete(f))
	return cmd
}
