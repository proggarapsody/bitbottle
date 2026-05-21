// Package create implements `bitbottle snippet create`.
package create

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	snippetlist "github.com/proggarapsody/bitbottle/pkg/cmd/snippet/list"
)

// Options carries parsed flags for `snippet create`.
type Options struct {
	Output    format.OutputConfig
	Hostname  string
	Workspace string
	Title     string
	Files     []string
	Private   bool
}

// NewCmdCreate constructs the cobra command. The runF override lets tests inject
// their own runner without standing up a real backend.
func NewCmdCreate(f *factory.Factory, runF ...func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a snippet from local files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if len(runF) > 0 && runF[0] != nil {
				return runF[0](opts)
			}
			return createRun(f, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Title, "title", "t", "", "Snippet title (required)")
	cmd.Flags().StringArrayVar(&opts.Files, "file", nil, "Path to file to include (repeatable)")
	cmd.Flags().BoolVar(&opts.Private, "private", false, "Make snippet private")
	cmd.Flags().StringVar(&opts.Workspace, "workspace", "", "Workspace slug (defaults to authenticated user)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func createRun(f *factory.Factory, opts *Options) error {
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

	snippetFiles, err := readFiles(opts.Files)
	if err != nil {
		return err
	}

	s, err := sc.CreateSnippet(workspace, backend.CreateSnippetInput{
		Title:     opts.Title,
		IsPrivate: opts.Private,
		Files:     snippetFiles,
	})
	if err != nil {
		return err
	}
	p := snippetlist.SnippetListFields(f, opts.Output)
	p.SetSingleItem()
	p.AddItem(s)
	return p.Render()
}

// readFiles reads each path into a SnippetFile. Empty file list is allowed
// (snippet with no files is valid per the API).
func readFiles(paths []string) ([]backend.SnippetFile, error) {
	out := make([]backend.SnippetFile, 0, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		out = append(out, backend.SnippetFile{
			Name:    filepath.Base(p),
			Content: string(data),
		})
	}
	return out, nil
}
