package download

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

func NewCmdDownloadList(f *factory.Factory) *cobra.Command {
	var hostname string
	var limit int

	cmd := &cobra.Command{
		Use:   "list [WORKSPACE/REPO]",
		Short: "List repository download artifacts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(limit); err != nil {
				return err
			}
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			rc, err := backend.AsRepoDownloadClient(client, ref.Host)
			if err != nil {
				return err
			}
			downloads, err := rc.ListRepoDownloads(ref.Project, ref.Slug, limit)
			if err != nil {
				return err
			}
			p := repoDownloadListFields(f, format.ConfigFromCmd(cmd))
			for _, d := range downloads {
				p.AddItem(d)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of downloads")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func repoDownloadListFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.RepoDownload] {
	p := format.New[backend.RepoDownload](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.RepoDownload]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(d backend.RepoDownload) any { return d.Name },
	})
	p.AddField(format.Field[backend.RepoDownload]{
		Name:   "size",
		Header: "SIZE",
		Extract: func(d backend.RepoDownload) any {
			return formatBytes(d.Size)
		},
	})
	p.AddField(format.Field[backend.RepoDownload]{
		Name:    "downloads",
		Header:  "DOWNLOADS",
		Extract: func(d backend.RepoDownload) any { return d.Downloads },
	})
	p.AddField(format.Field[backend.RepoDownload]{
		Name:    "createdOn",
		Header:  "CREATED",
		Extract: func(d backend.RepoDownload) any { return d.CreatedOn.Format("2006-01-02") },
	})
	return p
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
