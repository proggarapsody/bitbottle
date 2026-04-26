package commit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCommitView(f *factory.Factory) *cobra.Command {
	var web bool
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "view PROJECT/REPO HASH",
		Short: "View a commit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := f.ResolveRef(args[0], hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			hash := args[1]
			c, err := client.GetCommit(ref.Project, ref.Slug, hash)
			if err != nil {
				return err
			}

			if web {
				if c.WebURL == "" {
					return fmt.Errorf("no web URL available for this commit")
				}
				return f.Browser.Browse(c.WebURL)
			}

			if jsonFields != "" || jqExpr != "" {
				p := commitViewFields(f, jsonFields, jqExpr)
				p.SetSingleItem()
				p.AddItem(c)
				return p.Render()
			}

			out := f.IOStreams.Out
			fmt.Fprintf(out, "commit %s\n", c.Hash)
			fmt.Fprintf(out, "\n%s\n", c.Message)
			fmt.Fprintf(out, "\nAuthor:  %s\n", authorDisplay(c))
			fmt.Fprintf(out, "Date:    %s\n", c.Timestamp.UTC().Format("2006-01-02 15:04:05 +0000 UTC"))
			fmt.Fprintf(out, "Web:     %s\n", c.WebURL)
			return nil
		},
	}

	cmd.Flags().BoolVar(&web, "web", false, "Open commit in browser")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func commitViewFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Commit] {
	p := format.New[backend.Commit](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)

	p.AddField(format.Field[backend.Commit]{
		Name:    "hash",
		Header:  "HASH",
		Extract: func(c backend.Commit) any { return c.Hash },
	})
	p.AddField(format.Field[backend.Commit]{
		Name:    "message",
		Header:  "MESSAGE",
		Extract: func(c backend.Commit) any { return c.Message },
	})
	p.AddField(format.Field[backend.Commit]{
		Name:    "author",
		Header:  "AUTHOR",
		Extract: func(c backend.Commit) any { return authorDisplay(c) },
	})
	p.AddField(format.Field[backend.Commit]{
		Name:    "timestamp",
		Header:  "TIMESTAMP",
		Extract: func(c backend.Commit) any { return c.Timestamp.UTC().Format("2006-01-02T15:04:05Z") },
	})
	p.AddField(format.Field[backend.Commit]{
		Name:    "web_url",
		Header:  "WEB_URL",
		Extract: func(c backend.Commit) any { return c.WebURL },
	})

	return p
}

func authorDisplay(c backend.Commit) string {
	if c.Author.Slug != "" {
		return c.Author.Slug
	}
	return c.Author.DisplayName
}
