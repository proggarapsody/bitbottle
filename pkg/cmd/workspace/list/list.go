// Package list implements `bitbottle workspace list`.
package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options carries parsed flags for `workspace list`.
type Options struct {
	Output   format.OutputConfig
	Hostname string
	Limit    int
}

// NewCmdList constructs the cobra command. The runF parameter follows the
// gh-style override pattern: tests inject their own runner; production
// passes nil and gets the default listRun.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspaces the authenticated user belongs to",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of workspaces (0 = no cap)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	// Workspaces are an account-level resource — no repository context.
	// ResolveHost only consults config: it picks the single configured host,
	// or surfaces a typed "specify hostname" error when ambiguous. That's
	// the contract we want here; ResolveTarget would needlessly try to read
	// `git remote` and fail outside a checkout.
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}

	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	wc, err := backend.AsWorkspaceClient(client, host)
	if err != nil {
		return err
	}
	workspaces, err := wc.ListWorkspaces(opts.Limit)
	if err != nil {
		return err
	}

	p := workspaceFields(f, opts.Output)
	for _, w := range workspaces {
		p.AddItem(w)
	}
	return p.Render()
}

// workspaceFields wires the format printer for both TTY and JSON paths.
// UUID is JSON-only; it's noisy in a TTY column but useful for scripting.
func workspaceFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Workspace] {
	p := format.New[backend.Workspace](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Workspace]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(w backend.Workspace) any { return w.UUID }})
	p.AddField(format.Field[backend.Workspace]{Name: "slug", Header: "SLUG", Extract: func(w backend.Workspace) any { return w.Slug }})
	p.AddField(format.Field[backend.Workspace]{Name: "name", Header: "NAME", Extract: func(w backend.Workspace) any { return w.Name }})
	p.AddField(format.Field[backend.Workspace]{Name: "webURL", Header: "URL", Aliases: []string{"url", "link"}, Extract: func(w backend.Workspace) any { return w.WebURL }})
	return p
}
