package pipeline

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPipelineView(f *factory.Factory) *cobra.Command {
	var web bool
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "view PROJECT/REPO UUID",
		Short: "View a pipeline",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := resolvePipelineRef(f, args[0], hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			pc, ok := client.(backend.PipelineClient)
			if !ok {
				return fmt.Errorf("pipelines are only supported on Bitbucket Cloud")
			}

			pl, err := pc.GetPipeline(ref.Project, ref.Slug, args[1])
			if err != nil {
				return err
			}

			if web {
				if pl.WebURL == "" {
					return fmt.Errorf("no web URL available for this pipeline")
				}
				return f.Browser.Browse(pl.WebURL)
			}

			if jsonFields != "" || jqExpr != "" {
				printer := pipelineFields(f, jsonFields, jqExpr)
				printer.SetSingleItem()
				printer.AddItem(pl)
				return printer.Render()
			}

			out := f.IOStreams.Out
			fmt.Fprintf(out, "#%d %s\n", pl.BuildNumber, pl.State)
			fmt.Fprintf(out, "Branch:   %s\n", pl.RefName)
			fmt.Fprintf(out, "Duration: %ds\n", pl.Duration)
			if pl.WebURL != "" {
				fmt.Fprintf(out, "URL:      %s\n", pl.WebURL)
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
