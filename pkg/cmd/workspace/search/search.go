// Package search implements `bitbottle workspace search`.
package search

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// Options carries parsed flags for `workspace search`.
type Options struct {
	Output   format.OutputConfig
	Hostname string
	Query    string
	Role     string
	Limit    int
}

// NewCmdSearch constructs the cobra command. The runF parameter follows the
// gh-style override pattern: tests inject their own runner; production passes
// nil and gets the default searchRun.
func NewCmdSearch(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search workspaces by slug/name with optional role filter (Cloud only)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return searchRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Query, "query", "", "Slug/name prefix to match")
	cmd.Flags().StringVar(&opts.Role, "role", "", "Filter by role: owner, collaborator, or member")
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of workspaces (0 = no cap)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func searchRun(f *factory.Factory, opts *Options) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	wc, err := backend.AsWorkspaceSearcher(client, host)
	if err != nil {
		return err
	}
	workspaces, searchErr := wc.SearchWorkspaces(backend.WorkspaceSearchOpts{
		Query: opts.Query,
		Role:  opts.Role,
		Limit: opts.Limit,
	})
	if searchErr != nil && len(workspaces) == 0 {
		return searchErr
	}

	p := workspaceFields(f, opts.Output)
	for _, w := range workspaces {
		p.AddItem(w)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(workspaces), searchErr)
	return searchErr
}

// workspaceFields wires the format printer for both TTY and JSON paths.
func workspaceFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Workspace] {
	p := format.New[backend.Workspace](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Workspace]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(w backend.Workspace) any { return w.UUID }})
	p.AddField(format.Field[backend.Workspace]{Name: "slug", Header: "SLUG", Extract: func(w backend.Workspace) any { return w.Slug }})
	p.AddField(format.Field[backend.Workspace]{Name: "name", Header: "NAME", Extract: func(w backend.Workspace) any { return w.Name }})
	p.AddField(format.Field[backend.Workspace]{Name: "webURL", Header: "URL", Aliases: []string{"url", "link"}, Extract: func(w backend.Workspace) any { return w.WebURL }})
	return p
}
