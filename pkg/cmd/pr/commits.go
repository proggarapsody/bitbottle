package pr

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// NewCmdPRCommits returns the "pr commits" sub-command.
func NewCmdPRCommits(f *factory.Factory) *cobra.Command {
	var hostnameFlag string

	cmd := &cobra.Command{
		Use:         "commits PR_ID",
		Short:       "List commits in a pull request",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{cmdutil.PagerAnnotation: "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolvePRTarget(f, args, hostnameFlag)
			if err != nil {
				return err
			}

			pc, err := backend.AsPRCommitClient(client, ref.Host)
			if err != nil {
				return err
			}

			commits, err := pc.ListPRCommits(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			p := prCommitsFields(f, format.ConfigFromCmd(cmd))
			for _, c := range commits {
				p.AddItem(c)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostnameFlag, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func prCommitsFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Commit] {
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
		Name:   "message",
		Header: "MESSAGE",
		Extract: func(c backend.Commit) any {
			if isTTY && len(c.Message) > 60 {
				return c.Message[:60]
			}
			return c.Message
		},
	})

	p.AddField(format.Field[backend.Commit]{
		Name:   "author",
		Header: "AUTHOR",
		Extract: func(c backend.Commit) any {
			if c.Author.DisplayName != "" {
				return c.Author.DisplayName
			}
			return c.Author.Slug
		},
	})

	p.AddField(format.Field[backend.Commit]{
		Name:   "date",
		Header: "DATE",
		Extract: func(c backend.Commit) any {
			if structured || !isTTY {
				return c.Timestamp.Format(time.RFC3339)
			}
			return prCommitHumanizeTime(c.Timestamp)
		},
	})

	return p
}

func prCommitHumanizeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < 2*time.Minute:
		return "just now"
	case d < time.Hour:
		return "minutes ago"
	case d < 24*time.Hour:
		return "hours ago"
	case d < 48*time.Hour:
		return "yesterday"
	default:
		return t.Format("2006-01-02")
	}
}
