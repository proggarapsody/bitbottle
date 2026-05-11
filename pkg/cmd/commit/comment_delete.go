package commit

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdCommitCommentDelete removes an existing commit comment.
func NewCmdCommitCommentDelete(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO HASH COMMENT_ID",
		Short: "Delete an existing commit comment",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			commentID, err := strconv.Atoi(args[2])
			if err != nil || commentID <= 0 {
				return fmt.Errorf("invalid COMMENT_ID %q: must be a positive integer", args[2])
			}
			ref, err := factory.ResolveTarget(f, args[:1], hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			hash := args[1]
			return client.DeleteCommitComment(ref.Project, ref.Slug, hash, commentID)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
