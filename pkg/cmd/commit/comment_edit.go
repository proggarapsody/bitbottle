package commit

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdCommitCommentEdit updates the body of an existing commit comment.
func NewCmdCommitCommentEdit(f *factory.Factory) *cobra.Command {
	var body, hostname string

	cmd := &cobra.Command{
		Use:   "edit PROJECT/REPO HASH COMMENT_ID",
		Short: "Edit an existing commit comment",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
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
			if _, err := client.EditCommitComment(ref.Project, ref.Slug, hash, commentID, body); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "New comment body (required)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
