// Package view implements `bitbottle snippet view`.
package view

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	snippetlist "github.com/proggarapsody/bitbottle/pkg/cmd/snippet/list"
)

// Options carries parsed flags for `snippet view`.
type Options struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
	ID        string
}

// NewCmdView constructs the cobra command. The runF override lets tests inject
// their own runner without standing up a real backend.
func NewCmdView(f *factory.Factory, runF ...func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "view SNIPPET_ID",
		Short: "View a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ID = args[0]
			opts.Output = format.ConfigFromCmd(cmd)
			if len(runF) > 0 && runF[0] != nil {
				return runF[0](opts)
			}
			return viewRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Workspace, "workspace", "", "Workspace slug (defaults to authenticated user)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func viewRun(f *factory.Factory, opts *Options) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	workspace, err := snippetlist.ResolveWorkspace(f, host, opts.Workspace)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	sc, err := backend.AsSnippetClient(client, host)
	if err != nil {
		return err
	}
	s, err := sc.GetSnippet(workspace, opts.ID)
	if err != nil {
		return err
	}
	p := snippetViewFields(f, opts.Output)
	p.SetSingleItem()
	p.AddItem(s)
	return p.Render()
}

func snippetViewFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Snippet] {
	p := snippetlist.SnippetListFields(f, cfg)
	p.AddField(format.Field[backend.Snippet]{
		Name:   "created_on",
		Header: "CREATED",
		Extract: func(s backend.Snippet) any {
			if s.CreatedOn.IsZero() {
				return ""
			}
			return s.CreatedOn.Format("2006-01-02")
		},
	})
	return p
}
