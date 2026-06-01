package commit

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

func NewCmdCommitSearch(f *factory.Factory) *cobra.Command {
	var query, author, since, until string
	var limit int
	var hostname string

	cmd := &cobra.Command{
		Use:   "search [PROJECT/REPO]",
		Short: "Search commits by message, author, or date",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(limit); err != nil {
				return err
			}
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			searcher, err := backend.AsCommitSearcher(client, ref.Host)
			if err != nil {
				return err
			}

			commits, err := searcher.SearchCommits(ref.Project, ref.Slug, backend.CommitSearchOpts{
				Query:  query,
				Author: author,
				Since:  since,
				Until:  until,
				Limit:  limit,
			})
			if err != nil {
				return err
			}

			p := commitSearchFields(f, format.ConfigFromCmd(cmd))
			for _, c := range commits {
				p.AddItem(c)
			}
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Filter by message keyword")
	cmd.Flags().StringVar(&author, "author", "", "Filter by author slug or display name")
	cmd.Flags().StringVar(&since, "since", "", "Start date or commit SHA (ISO 8601)")
	cmd.Flags().StringVar(&until, "until", "", "End date or commit SHA (ISO 8601)")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of commits")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func commitSearchFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Commit] {
	isTTY := f.IOStreams.IsStdoutTTY()
	structured := cfg.Format != format.FormatTable
	p := format.New[backend.Commit](f.IOStreams.Out, isTTY, cfg)

	p.AddField(format.Field[backend.Commit]{
		Name:   "hash",
		Header: "HASH",
		Extract: func(c backend.Commit) any {
			if isTTY && len(c.Hash) >= 8 {
				return c.Hash[:8]
			}
			return c.Hash
		},
	})

	p.AddField(format.Field[backend.Commit]{
		Name:   "author",
		Header: "AUTHOR",
		Extract: func(c backend.Commit) any {
			return authorDisplay(c)
		},
	})

	p.AddField(format.Field[backend.Commit]{
		Name:   "date",
		Header: "DATE",
		Extract: func(c backend.Commit) any {
			if structured || !isTTY {
				return c.Timestamp.Format(time.RFC3339)
			}
			return humanizeTime(c.Timestamp)
		},
	})

	p.AddField(format.Field[backend.Commit]{
		Name:   "message",
		Header: "MESSAGE",
		Extract: func(c backend.Commit) any {
			msg := c.Message
			if isTTY && len(msg) > 60 {
				return msg[:60]
			}
			return msg
		},
	})

	return p
}
