package pipeline

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPipelineList(f *factory.Factory) *cobra.Command {
	var limit int
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List pipelines",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := f.ResolveRef(args[0], hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			pc, err := backend.AsPipelineClient(client)
			if err != nil {
				return err
			}

			pipelines, err := pc.ListPipelines(ref.Project, ref.Slug, limit)
			if err != nil {
				return err
			}

			p := pipelineFields(f, jsonFields, jqExpr)
			for _, pl := range pipelines {
				p.AddItem(pl)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of pipelines")
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func pipelineFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Pipeline] {
	p := format.New[backend.Pipeline](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Pipeline]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(pl backend.Pipeline) any { return pl.UUID }})
	p.AddField(format.Field[backend.Pipeline]{Name: "buildNumber", Header: "BUILD", Extract: func(pl backend.Pipeline) any { return pl.BuildNumber }})
	p.AddField(format.Field[backend.Pipeline]{Name: "state", Header: "STATE", Extract: func(pl backend.Pipeline) any { return pl.State }})
	p.AddField(format.Field[backend.Pipeline]{Name: "refName", Header: "BRANCH/TAG", Extract: func(pl backend.Pipeline) any { return pl.RefName }})
	p.AddField(format.Field[backend.Pipeline]{Name: "duration", Header: "DURATION", Extract: func(pl backend.Pipeline) any { return pl.Duration }})
	p.AddField(format.Field[backend.Pipeline]{Name: "webURL", Header: "URL", JSONOnly: true, Extract: func(pl backend.Pipeline) any { return pl.WebURL }})
	return p
}
