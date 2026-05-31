package commit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// NewCmdCommitCommentAdd posts a new comment on a commit.
func NewCmdCommitCommentAdd(f *factory.Factory) *cobra.Command {
	var body, hostname string

	cmd := &cobra.Command{
		Use:   "add [PROJECT/REPO] HASH",
		Short: "Add a comment to a commit",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			repoArgs, rest := repoarg.SplitLeadingRepo(args, 1)
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			hash := rest[0]
			c, err := client.AddCommitComment(ref.Project, ref.Slug, hash, backend.AddCommitCommentInput{Body: body})
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Created comment %d\n", c.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "Comment body (required)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
