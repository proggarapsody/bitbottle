// Package mirror implements the `bitbottle mirror` command group.
// Smart Mirror is a Bitbucket Server/DC feature; the optional MirrorClient
// interface gates these commands so an invocation against a Cloud host
// surfaces a typed ErrUnsupportedOnHost rather than a runtime panic.
package mirror

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdMirror)
}

// NewCmdMirror returns the top-level `mirror` command.
func NewCmdMirror(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mirror",
		Short: "Manage Smart Mirror servers (Server/DC only)",
	}
	cmd.AddCommand(NewCmdMirrorList(f, nil))
	cmd.AddCommand(NewCmdMirrorView(f, nil))
	cmd.AddCommand(NewCmdMirrorRepo(f))
	return cmd
}
