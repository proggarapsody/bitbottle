package commit

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// NewCmdCommitCommentList lists all comments on a commit.
func NewCmdCommitCommentList(f *factory.Factory) *cobra.Command {
	var hostname string
	var withReactions bool

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
			cmts, listErr := client.ListCommitComments(ref.Project, ref.Slug, hash, 0)
			if listErr != nil && len(cmts) == 0 {
				return listErr
			}
			if withReactions {
				reactor, reactErr := backend.AsCommitCommentReactor(client, ref.Host)
				if reactErr != nil {
					return reactErr
				}
				cmts = fetchCommitCommentReactionsConcurrent(reactor, ref.Project, ref.Slug, hash, cmts)
			}
			p := commitCommentFields(f, format.ConfigFromCmd(cmd), withReactions)
			for _, c := range cmts {
				p.AddItem(c)
			}
			if err := p.Render(); err != nil {
				return err
			}
			cmdutil.PartialWarn(f.IOStreams.ErrOut, len(cmts), listErr)
			return listErr
		},
	}
	cmd.Flags().BoolVar(&withReactions, "reactions", false, "Fetch and display emoji reactions (Bitbucket Server / DC only)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// fetchCommitCommentReactionsConcurrent fetches reactions for each commit
// comment concurrently using a bounded worker pool of 4 goroutines. The
// returned slice has the same order as the input; errors per-comment are
// silently ignored so a single failure doesn't abort the listing.
func fetchCommitCommentReactionsConcurrent(reactor backend.CommitCommentReactor, ns, slug, hash string, cmts []backend.CommitComment) []backend.CommitComment {
	const workers = 4
	type job struct {
		idx int
		id  int
	}
	results := make([][]backend.CommentReaction, len(cmts))
	jobs := make(chan job, len(cmts))
	for i, c := range cmts {
		jobs <- job{i, c.ID}
	}
	close(jobs)

	var wg sync.WaitGroup
	var mu sync.Mutex
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				rxns, err := reactor.ListCommitCommentReactions(ns, slug, hash, j.id)
				if err == nil && len(rxns) > 0 {
					mu.Lock()
					results[j.idx] = rxns
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	out := make([]backend.CommitComment, len(cmts))
	for i, c := range cmts {
		c.Reactions = results[i]
		out[i] = c
	}
	return out
}

// formatCommitReactions renders a CommentReaction slice as a compact string
// like "👍×3 ❤️×1". Returns "" when the slice is empty.
func formatCommitReactions(reactions []backend.CommentReaction) string {
	if len(reactions) == 0 {
		return ""
	}
	emojiGlyphs := map[string]string{
		"thumbs_up":   "👍",
		"thumbs_down": "👎",
		"heart":       "❤️",
		"laugh":       "😄",
		"hooray":      "🎉",
		"confused":    "😕",
	}
	parts := make([]string, 0, len(reactions))
	for _, r := range reactions {
		glyph := emojiGlyphs[r.Emoji]
		if glyph == "" {
			glyph = r.Emoji
		}
		parts = append(parts, fmt.Sprintf("%s×%d", glyph, len(r.Users)))
	}
	return strings.Join(parts, " ")
}

func commitCommentFields(f *factory.Factory, cfg format.OutputConfig, showReactions bool) *format.Printer[backend.CommitComment] {
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
	// reactions: shown as TTY column when --reactions is set, otherwise JSON-only.
	p.AddField(format.Field[backend.CommitComment]{
		Name:     "reactions",
		Header:   "REACTIONS",
		JSONOnly: !showReactions,
		Extract: func(c backend.CommitComment) any {
			if len(c.Reactions) == 0 {
				return nil
			}
			if structured {
				type reactionJSON struct {
					Emoji string   `json:"emoji"`
					Users []string `json:"users"`
				}
				out := make([]reactionJSON, 0, len(c.Reactions))
				for _, r := range c.Reactions {
					slugs := make([]string, 0, len(r.Users))
					for _, u := range r.Users {
						slugs = append(slugs, u.Slug)
					}
					out = append(out, reactionJSON{Emoji: r.Emoji, Users: slugs})
				}
				return out
			}
			return formatCommitReactions(c.Reactions)
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
