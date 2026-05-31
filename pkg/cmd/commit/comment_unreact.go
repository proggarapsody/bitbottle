package commit

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// NewCmdCommitCommentUnreact removes an emoji reaction from a commit comment.
func NewCmdCommitCommentUnreact(f *factory.Factory) *cobra.Command {
	var emoji, hostname string

	cmd := &cobra.Command{
		Use:   "unreact [PROJECT/REPO] HASH COMMENT_ID",
		Short: "Remove emoji reaction from commit comment (Server)",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if emoji == "" {
				return fmt.Errorf("--emoji is required")
			}
			repoArgs, rest := repoarg.SplitLeadingRepo(args, 2)
			commentID, err := strconv.Atoi(rest[1])
			if err != nil || commentID <= 0 {
				return fmt.Errorf("invalid COMMENT_ID %q: must be a positive integer", rest[1])
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			reactor, err := backend.AsCommitCommentReactor(client, ref.Host)
			if err != nil {
				return err
			}
			hash := rest[0]
			if err := reactor.RemoveCommitCommentReaction(ref.Project, ref.Slug, hash, commentID, backend.NormaliseEmoji(emoji)); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Removed reaction from comment #%d\n", commentID)
			return nil
		},
	}
	cmd.Flags().StringVar(&emoji, "emoji", "", "Emoji shortcode: thumbs_up, thumbs_down, heart, laugh, hooray, confused (required)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
