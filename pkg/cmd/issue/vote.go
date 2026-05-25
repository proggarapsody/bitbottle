package issue

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdIssueVote builds `issue vote [PROJECT/REPO] ISSUE_ID`.
func NewCmdIssueVote(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "vote [PROJECT/REPO] ISSUE_ID",
		Short: "Vote on a Bitbucket Cloud issue",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			iv, err := backend.AsIssueVoter(client, ref.Host)
			if err != nil {
				return err
			}
			if err := iv.VoteIssue(ref.Project, ref.Slug, id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Voted on issue #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// NewCmdIssueUnvote builds `issue unvote [PROJECT/REPO] ISSUE_ID`.
func NewCmdIssueUnvote(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "unvote [PROJECT/REPO] ISSUE_ID",
		Short: "Remove vote from a Bitbucket Cloud issue",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid issue ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			iv, err := backend.AsIssueVoter(client, ref.Host)
			if err != nil {
				return err
			}
			if err := iv.UnvoteIssue(ref.Project, ref.Slug, id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Vote removed from issue #%d\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
