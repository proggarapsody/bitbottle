package cache

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ListOptions holds parsed flags for `pipeline cache list`.
type ListOptions struct {
	Hostname string
	Args     []string
}

// NewCmdList builds the `pipeline cache list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List pipeline caches",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runList(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

// formatBytes renders a byte count as a human-readable string (e.g. "11.8 MB").
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func runList(f *factory.Factory, cmd *cobra.Command, opts *ListOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPipelineCacheClient(client, ref.Host)
	if err != nil {
		return err
	}
	caches, err := pc.ListPipelineCaches(ref.Project, ref.Slug)
	if err != nil {
		return err
	}

	cfg := format.ConfigFromCmd(cmd)
	if cfg.Format != format.FormatTable {
		p := format.New[backend.PipelineCache](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
		p.AddField(format.Field[backend.PipelineCache]{Name: "uuid", Header: "UUID", Extract: func(c backend.PipelineCache) any { return c.UUID }})
		p.AddField(format.Field[backend.PipelineCache]{Name: "name", Header: "NAME", Extract: func(c backend.PipelineCache) any { return c.Name }})
		p.AddField(format.Field[backend.PipelineCache]{Name: "path", Header: "PATH", Extract: func(c backend.PipelineCache) any { return c.Path }})
		p.AddField(format.Field[backend.PipelineCache]{Name: "fileSizeBytes", Header: "SIZE", Extract: func(c backend.PipelineCache) any { return c.FileSizeBytes }})
		p.AddField(format.Field[backend.PipelineCache]{Name: "createdOn", Header: "CREATED", Extract: func(c backend.PipelineCache) any { return c.CreatedOn }})
		for _, c := range caches {
			p.AddItem(c)
		}
		return p.Render()
	}

	out := f.IOStreams.Out
	if len(caches) == 0 {
		fmt.Fprintln(out, "No pipeline caches found.")
		return nil
	}
	fmt.Fprintf(out, "%-38s  %-20s  %-30s  %10s  %s\n", "UUID", "NAME", "PATH", "SIZE", "CREATED")
	for _, c := range caches {
		fmt.Fprintf(out, "%-38s  %-20s  %-30s  %10s  %s\n",
			c.UUID, c.Name, c.Path, formatBytes(c.FileSizeBytes), c.CreatedOn)
	}
	return nil
}
