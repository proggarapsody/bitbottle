// Package list implements `bitbottle snippet list`.
package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// Options carries parsed flags for `snippet list`.
type Options struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
	Limit     int
}

// NewCmdList constructs the cobra command. The runF override lets tests inject
// their own runner without standing up a real backend.
func NewCmdList(f *factory.Factory, runF ...func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List snippets in a workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if err := cmdutil.ValidatePositiveLimit(opts.Limit); err != nil {
				return err
			}
			if len(runF) > 0 && runF[0] != nil {
				return runF[0](opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Workspace, "workspace", "", "Workspace slug (defaults to authenticated user)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of snippets (0 = no cap)")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	workspace, err := ResolveWorkspace(f, host, opts.Workspace)
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
	snippets, err := sc.ListSnippets(workspace, opts.Limit)
	if err != nil {
		return err
	}
	p := SnippetListFields(f, opts.Output)
	for _, s := range snippets {
		p.AddItem(s)
	}
	return p.Render()
}

// ResolveWorkspace returns the workspace to use: explicit flag wins, then the
// User slug from the host's config entry (the Bitbucket username/workspace slug).
// Exported so sibling packages (view, create, delete) can share the logic.
func ResolveWorkspace(f *factory.Factory, host, flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	hc, _ := cfg.Get(host)
	if hc.User != "" {
		return hc.User, nil
	}
	return "", fmt.Errorf("cannot determine workspace: specify --workspace or configure a user for host %s", host)
}

// SnippetListFields builds a printer with the standard snippet list fields.
// Exported so sibling commands (view, etc.) can extend it.
func SnippetListFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Snippet] {
	p := format.New[backend.Snippet](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Snippet]{Name: "id", Header: "ID", Extract: func(s backend.Snippet) any { return s.ID }})
	p.AddField(format.Field[backend.Snippet]{Name: "title", Header: "TITLE", Extract: func(s backend.Snippet) any { return s.Title }})
	p.AddField(format.Field[backend.Snippet]{Name: "owner", Header: "OWNER", Extract: func(s backend.Snippet) any { return s.Owner }})
	p.AddField(format.Field[backend.Snippet]{Name: "private", Header: "PRIVATE", Extract: func(s backend.Snippet) any { return s.IsPrivate }})
	p.AddField(format.Field[backend.Snippet]{Name: "files", Header: "FILES", Extract: func(s backend.Snippet) any { return len(s.Files) }})
	p.AddField(format.Field[backend.Snippet]{Name: "url", Header: "URL", Aliases: []string{"webURL", "link"}, Extract: func(s backend.Snippet) any { return s.WebURL }})
	return p
}
