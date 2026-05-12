package search

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdSearchCode returns the `bitbottle search code QUERY` command.
//
// Resolution order for the workspace:
//  1. --workspace flag (explicit win).
//  2. BaseRepo's Project (origin remote / pinned default / -R override).
//
// The query string is passed to Bitbucket Cloud verbatim — operators like
// `path:`, `lang:`, etc. are interpreted by Bitbucket itself.
func NewCmdSearchCode(f *factory.Factory) *cobra.Command {
	var workspace, hostname string
	var limit int
	cmd := &cobra.Command{
		Use:   "code QUERY",
		Short: "Search code across a Bitbucket Cloud workspace (Cloud only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]

			host, ws, err := resolveSearchTarget(f, workspace, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			cs, err := backend.AsCodeSearcher(client, host)
			if err != nil {
				return err
			}
			hits, err := cs.SearchCode(ws, query, limit)
			if err != nil {
				return err
			}
			p := codeSearchFields(f, format.ConfigFromCmd(cmd))
			for _, h := range hits {
				p.AddItem(h)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&workspace, "workspace", "", "Bitbucket Cloud workspace slug (defaults to current repo's workspace)")
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of results")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// resolveSearchTarget picks the (host, workspace) pair the search should
// hit. An explicit --workspace flag wins; otherwise the BaseRepo's Project
// supplies the workspace (and its Host disambiguates which configured
// Cloud the user means). --hostname overrides the resolved host.
func resolveSearchTarget(f *factory.Factory, workspaceFlag, hostnameFlag string) (host, workspace string, err error) {
	if workspaceFlag != "" {
		// With an explicit workspace, host inference becomes the standard
		// ResolveHost rule: --hostname wins, else single-host fallback,
		// else "specify hostname" error.
		host, err = factory.ResolveHost(f, hostnameFlag)
		if err != nil {
			return "", "", err
		}
		return host, workspaceFlag, nil
	}

	ref, err := f.BaseRepo()
	if err != nil {
		return "", "", fmt.Errorf("workspace required: pass --workspace or run inside a repository checkout")
	}
	host = ref.Host
	if hostnameFlag != "" {
		host = hostnameFlag
	}
	return host, ref.Project, nil
}

// codeSearchFields wires the Printer columns / JSON keys for CodeSearchHit.
// Matched-segment text is unrolled into "match" strings so JSON consumers
// can grep for the literal hits without re-parsing the segment shape.
func codeSearchFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.CodeSearchHit] {
	p := format.New[backend.CodeSearchHit](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.CodeSearchHit]{Name: "repository", Header: "REPOSITORY", Extract: func(h backend.CodeSearchHit) any { return h.Repository }})
	p.AddField(format.Field[backend.CodeSearchHit]{Name: "path", Header: "PATH", Extract: func(h backend.CodeSearchHit) any { return h.Path }})
	p.AddField(format.Field[backend.CodeSearchHit]{
		Name: "contentMatchCount", Header: "MATCHES",
		Aliases: []string{"matches", "content_match_count"},
		Extract: func(h backend.CodeSearchHit) any { return h.ContentMatchCount },
	})
	p.AddField(format.Field[backend.CodeSearchHit]{
		Name:     "url",
		Aliases:  []string{"webURL", "fileURL", "link"},
		Extract:  func(h backend.CodeSearchHit) any { return h.FileURL },
		JSONOnly: true,
	})
	p.AddField(format.Field[backend.CodeSearchHit]{
		Name:     "pathMatches",
		Aliases:  []string{"path_matches"},
		Extract:  func(h backend.CodeSearchHit) any { return h.PathMatches },
		JSONOnly: true,
	})
	p.AddField(format.Field[backend.CodeSearchHit]{
		Name:     "contentMatches",
		Aliases:  []string{"content_matches"},
		Extract:  func(h backend.CodeSearchHit) any { return h.ContentMatches },
		JSONOnly: true,
	})
	return p
}
