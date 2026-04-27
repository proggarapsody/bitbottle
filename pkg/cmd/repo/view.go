package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoView(f *factory.Factory) *cobra.Command {
	var web bool
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "view [PROJECT/REPO]",
		Short: "View a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			r, err := client.GetRepo(ref.Project, ref.Slug)
			if err != nil {
				return err
			}

			if web {
				if r.WebURL == "" {
					return fmt.Errorf("no web URL available for this repository")
				}
				return f.Browser.Browse(r.WebURL)
			}

			if jsonFields != "" || jqExpr != "" {
				p := repoFields(f, jsonFields, jqExpr)
				p.SetSingleItem()
				p.AddItem(r)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Name:      %s\n", r.Slug)
			fmt.Fprintf(f.IOStreams.Out, "Namespace: %s\n", r.Namespace)
			fmt.Fprintf(f.IOStreams.Out, "SCM:       %s\n", r.SCM)
			if r.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "URL:       %s\n", r.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&web, "web", false, "Open in browser")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func repoFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Repository] {
	p := format.New[backend.Repository](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Repository]{Name: "slug", Header: "SLUG", Extract: func(r backend.Repository) any { return r.Slug }})
	p.AddField(format.Field[backend.Repository]{Name: "name", Header: "NAME", Extract: func(r backend.Repository) any { return r.Name }})
	p.AddField(format.Field[backend.Repository]{Name: "namespace", Header: "PROJECT", Extract: func(r backend.Repository) any { return r.Namespace }})
	p.AddField(format.Field[backend.Repository]{Name: "scm", Header: "TYPE", Extract: func(r backend.Repository) any { return r.SCM }})
	p.AddField(format.Field[backend.Repository]{Name: "webURL", Header: "URL", Extract: func(r backend.Repository) any { return r.WebURL }})
	return p
}
