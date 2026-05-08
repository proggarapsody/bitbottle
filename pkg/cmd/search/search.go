// Package search implements the `bitbottle search` command group. Code
// search is a Bitbucket Cloud feature gated by the CodeSearcher optional
// interface; invocations against Server/DC surface a typed
// ErrUnsupportedOnHost error.
package search

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdSearch(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search across a Bitbucket Cloud workspace (Cloud only)",
	}
	cmd.AddCommand(NewCmdSearchCode(f))
	return cmd
}
