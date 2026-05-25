// Package issue implements the `bitbottle issue` command group. Issues are
// a Bitbucket Cloud feature gated by the IssueClient optional interface;
// invocations against Server/DC surface a typed ErrUnsupportedOnHost.
package issue

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdIssue(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage Bitbucket Cloud issues (Cloud only)",
		Annotations: map[string]string{
			"help:arguments": `Most issue commands accept a repository as WORKSPACE/REPO. When
omitted, the repository is inferred from the "origin" git remote in
the current directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdIssueList(f))
	cmd.AddCommand(NewCmdIssueView(f))
	cmd.AddCommand(NewCmdIssueCreate(f))
	cmd.AddCommand(NewCmdIssueClose(f))
	cmd.AddCommand(NewCmdIssueEdit(f))
	cmd.AddCommand(NewCmdIssueReopen(f))
	cmd.AddCommand(NewCmdIssueAssign(f))
	cmd.AddCommand(NewCmdIssueComment(f))
	cmd.AddCommand(NewCmdIssueAttachment(f))
	cmd.AddCommand(NewCmdIssueVote(f))
	cmd.AddCommand(NewCmdIssueUnvote(f))
	cmd.AddCommand(NewCmdIssueWatch(f))
	cmd.AddCommand(NewCmdIssueUnwatch(f))
	return cmd
}
