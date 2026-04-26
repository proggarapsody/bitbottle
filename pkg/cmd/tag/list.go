package tag

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdTagList(f *factory.Factory) *cobra.Command {
	var limit int
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List tags",
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

			tags, err := client.ListTags(ref.Project, ref.Slug, limit)
			if err != nil {
				return err
			}

			p := tagFields(f, jsonFields, jqExpr)
			for _, t := range tags {
				p.AddItem(t)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of tags")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func tagFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Tag] {
	p := format.New[backend.Tag](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Tag]{Name: "name", Header: "NAME", Extract: func(t backend.Tag) any { return t.Name }})
	p.AddField(format.Field[backend.Tag]{Name: "hash", Header: "HASH", Extract: func(t backend.Tag) any {
		if len(t.Hash) > 8 {
			return t.Hash[:8]
		}
		return t.Hash
	}})
	p.AddField(format.Field[backend.Tag]{Name: "message", Header: "MESSAGE", Extract: func(t backend.Tag) any {
		if t.Message == "" {
			return ""
		}
		lines := strings.SplitN(t.Message, "\n", 2)
		return lines[0]
	}})
	return p
}
