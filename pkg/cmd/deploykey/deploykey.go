// Package deploykey implements the `bitbottle deploy-key` command group.
// Deploy keys are SSH public keys that grant read (or read/write) access to a
// repository. Both Bitbucket Cloud and Server/DC support deploy keys.
package deploykey

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdDeployKey)
}

// NewCmdDeployKey builds the root `deploy-key` command.
func NewCmdDeployKey(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-key",
		Short: "Manage repository deploy keys",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO (Server) or
WORKSPACE/REPO (Cloud). When omitted, the repository is inferred from
the "origin" git remote in the current directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdList(f))
	cmd.AddCommand(NewCmdAdd(f))
	cmd.AddCommand(NewCmdDelete(f))
	return cmd
}
