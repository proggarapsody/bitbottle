package tag

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

func NewCmdTagList(f *factory.Factory) *cobra.Command {
	var limit int
	var web bool
	var hostname string

	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List tags",
		Args:  cobra.ExactArgs(1),
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

			if web {
				repo, err := client.GetRepo(ref.Project, ref.Slug)
				if err != nil {
					return err
				}
				if repo.WebURL == "" {
					return fmt.Errorf("no web URL available for this repository")
				}
				return f.Browser.Browse(repo.WebURL)
			}

			tags, err := client.ListTags(ref.Project, ref.Slug, limit)
			if err != nil {
				return err
			}

			p := tagFields(f, format.ConfigFromCmd(cmd))
			for _, t := range tags {
				p.AddItem(t)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of tags")
	cmd.Flags().BoolVar(&web, "web", false, "Open repository tags page in browser")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func tagFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Tag] {
	p := format.New[backend.Tag](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
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
