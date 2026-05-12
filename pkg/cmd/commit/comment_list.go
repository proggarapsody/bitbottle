package commit

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdCommitCommentList lists all comments on a commit.
func NewCmdCommitCommentList(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO HASH",
		Short: "List all comments on a commit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args[:1], hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			hash := args[1]
			cmts, err := client.ListCommitComments(ref.Project, ref.Slug, hash, 0)
			if err != nil {
				return err
			}
			p := commitCommentFields(f, format.ConfigFromCmd(cmd))
			for _, c := range cmts {
				p.AddItem(c)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func commitCommentFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.CommitComment] {
	isTTY := f.IOStreams.IsStdoutTTY()
	structured := cfg.Format != format.FormatTable
	p := format.New[backend.CommitComment](f.IOStreams.Out, isTTY, cfg)
	p.AddField(format.Field[backend.CommitComment]{
		Name:    "id",
		Header:  "ID",
		Extract: func(c backend.CommitComment) any { return c.ID },
	})
	p.AddField(format.Field[backend.CommitComment]{
		Name:   "author",
		Header: "AUTHOR",
		Extract: func(c backend.CommitComment) any {
			if c.Author.Slug != "" {
				return c.Author.Slug
			}
			return c.Author.DisplayName
		},
	})
	p.AddField(format.Field[backend.CommitComment]{
		Name:   "createdAt",
		Header: "CREATED",
		Extract: func(c backend.CommitComment) any {
			if structured || !isTTY {
				return c.CreatedAt.Format(time.RFC3339)
			}
			return c.CreatedAt.Format("2006-01-02 15:04")
		},
	})
	p.AddField(format.Field[backend.CommitComment]{
		Name:   "body",
		Header: "BODY",
		Extract: func(c backend.CommitComment) any {
			if isTTY && len(c.Body) > 70 {
				return c.Body[:70]
			}
			return c.Body
		},
	})
	// JSON-only fields
	p.AddField(format.Field[backend.CommitComment]{
		Name:     "updatedAt",
		Header:   "UPDATED",
		JSONOnly: true,
		Extract: func(c backend.CommitComment) any {
			if c.UpdatedAt.IsZero() {
				return nil
			}
			return c.UpdatedAt.Format(time.RFC3339)
		},
	})
	return p
}
