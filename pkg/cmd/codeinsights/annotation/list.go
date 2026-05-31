package annotation

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// ListOptions holds parsed flags for `code-insights annotation list`.
type ListOptions struct {
	Output   format.OutputConfig
	Hostname string
	// Args[0]=PROJECT/REPO  Args[1]=HASH  Args[2]=KEY
	Args []string
}

// NewCmdList builds the `code-insights annotation list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO] HASH KEY",
		Short: "List Code Insights annotations for a report",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func annotationFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.CodeInsightsAnnotation] {
	p := format.New[backend.CodeInsightsAnnotation](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.CodeInsightsAnnotation]{
		Name: "path", Header: "PATH",
		Extract: func(a backend.CodeInsightsAnnotation) any { return a.Path },
	})
	p.AddField(format.Field[backend.CodeInsightsAnnotation]{
		Name: "line", Header: "LINE",
		Extract: func(a backend.CodeInsightsAnnotation) any { return a.Line },
	})
	p.AddField(format.Field[backend.CodeInsightsAnnotation]{
		Name: "severity", Header: "SEVERITY",
		Extract: func(a backend.CodeInsightsAnnotation) any { return a.Severity },
	})
	p.AddField(format.Field[backend.CodeInsightsAnnotation]{
		Name: "type", Header: "TYPE",
		Extract: func(a backend.CodeInsightsAnnotation) any { return a.Type },
	})
	p.AddField(format.Field[backend.CodeInsightsAnnotation]{
		Name: "message", Header: "MESSAGE",
		Extract: func(a backend.CodeInsightsAnnotation) any { return a.Message },
	})
	p.AddField(format.Field[backend.CodeInsightsAnnotation]{
		Name: "external_id", Header: "EXTERNAL_ID",
		JSONOnly: true,
		Extract:  func(a backend.CodeInsightsAnnotation) any { return a.ExternalID },
	})
	p.AddField(format.Field[backend.CodeInsightsAnnotation]{
		Name: "link", Header: "LINK",
		JSONOnly: true,
		Extract:  func(a backend.CodeInsightsAnnotation) any { return a.Link },
	})
	return p
}

func listRun(f *factory.Factory, opts *ListOptions) error {
	repoArgs, rest := repoarg.SplitLeadingRepo(opts.Args, 2)
	ref, err := factory.ResolveTarget(f, repoArgs, opts.Hostname)
	if err != nil {
		return err
	}
	hash := rest[0]
	key := rest[1]
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	ci, err := backend.AsCodeInsightsClient(client, ref.Host)
	if err != nil {
		return err
	}
	anns, err := ci.ListAnnotations(ref.Project, ref.Slug, hash, key)
	if err != nil {
		return err
	}
	p := annotationFields(f, opts.Output)
	for _, a := range anns {
		p.AddItem(a)
	}
	return p.Render()
}
