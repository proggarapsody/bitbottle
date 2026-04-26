package commit

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCommitLog(f *factory.Factory) *cobra.Command {
	var branch string
	var limit int
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "log PROJECT/REPO",
		Short: "List commits",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := f.ResolveRef(args[0], hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			if branch == "" {
				branch = currentGitBranch()
			}

			commits, err := client.ListCommits(ref.Project, ref.Slug, branch, limit)
			if err != nil {
				return err
			}

			p := commitLogFields(f, jsonFields, jqExpr)
			for _, c := range commits {
				p.AddItem(c)
			}
			return p.Render()
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch name (defaults to current git branch)")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of commits")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func commitLogFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Commit] {
	p := format.New[backend.Commit](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)

	p.AddField(format.Field[backend.Commit]{
		Name:   "hash",
		Header: "HASH",
		Extract: func(c backend.Commit) any {
			if f.IOStreams.IsStdoutTTY() {
				if len(c.Hash) >= 7 {
					return c.Hash[:7]
				}
				return c.Hash
			}
			return c.Hash
		},
	})

	p.AddField(format.Field[backend.Commit]{
		Name:   "message",
		Header: "MESSAGE",
		Extract: func(c backend.Commit) any {
			if f.IOStreams.IsStdoutTTY() {
				if len(c.Message) > 60 {
					return c.Message[:60]
				}
				return c.Message
			}
			return c.Message
		},
	})

	p.AddField(format.Field[backend.Commit]{
		Name:   "author",
		Header: "AUTHOR",
		Extract: func(c backend.Commit) any {
			if c.Author.Slug != "" {
				return c.Author.Slug
			}
			return c.Author.DisplayName
		},
	})

	p.AddField(format.Field[backend.Commit]{
		Name:   "date",
		Header: "DATE",
		Extract: func(c backend.Commit) any {
			if jsonFields != "" {
				return c.Timestamp.Format(time.RFC3339)
			}
			if f.IOStreams.IsStdoutTTY() {
				return humanizeTime(c.Timestamp)
			}
			return c.Timestamp.Format(time.RFC3339)
		},
	})

	return p
}

func currentGitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(out)) == "HEAD" {
		return "main"
	}
	return strings.TrimSpace(string(out))
}

func humanizeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < 2*time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}
