// Package snippet implements the `bitbottle snippet` command group.
// Snippets are a Bitbucket Cloud feature (gist parity); they are
// gated behind AsSnippetClient which returns ErrUnsupportedOnHost
// for Server/DC backends.
package snippet

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/comment"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/create"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/delete"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/list"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/view"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdSnippet)
}

// NewCmdSnippet builds the root `snippet` command.
func NewCmdSnippet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snippet",
		Short: "Manage Bitbucket Cloud snippets",
	}
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(view.NewCmdView(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(delete.NewCmdDelete(f))
	cmd.AddCommand(comment.NewCmdSnippetComment(f))
	return cmd
}
