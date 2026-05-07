// Package list implements `bitbottle project list WORKSPACE`.
package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options carries parsed flags + the WORKSPACE positional arg.
type Options struct {
	Hostname   string
	JSONFields string
	JQExpr     string
	Limit      int

	// Args[0] = WORKSPACE
	Args []string
}

func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list WORKSPACE",
		Short: "List projects within a Bitbucket Cloud workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of projects (0 = no cap)")
	cmd.Flags().StringVar(&opts.JSONFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&opts.JQExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Cloud only)")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	workspace := opts.Args[0]

	// Same rationale as workspace/list: project listing is workspace-scoped,
	// not repo-scoped, so ResolveHost (config-only) is the right resolver.
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
	projects, err := wc.ListProjects(workspace, opts.Limit)
	if err != nil {
		return err
	}

	p := projectFields(f, opts.JSONFields, opts.JQExpr)
	for _, pr := range projects {
		p.AddItem(pr)
	}
	return p.Render()
}

func projectFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Project] {
	p := format.New[backend.Project](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Project]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(pr backend.Project) any { return pr.UUID }})
	p.AddField(format.Field[backend.Project]{Name: "key", Header: "KEY", Extract: func(pr backend.Project) any { return pr.Key }})
	p.AddField(format.Field[backend.Project]{Name: "name", Header: "NAME", Extract: func(pr backend.Project) any { return pr.Name }})
	p.AddField(format.Field[backend.Project]{Name: "webURL", Header: "URL", Aliases: []string{"url", "link"}, Extract: func(pr backend.Project) any { return pr.WebURL }})
	return p
}
