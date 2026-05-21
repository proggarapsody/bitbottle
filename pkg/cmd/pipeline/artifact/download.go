package artifact

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DownloadOptions holds parsed flags for `pipeline artifact download`.
type DownloadOptions struct {
	Hostname     string
	PipelineUUID string
	StepUUID     string
	Name         string
	Out          string
	Args         []string
}

// NewCmdDownload builds the `pipeline artifact download` cobra command.
func NewCmdDownload(f *factory.Factory, runF func(*DownloadOptions) error) *cobra.Command {
	opts := &DownloadOptions{}
	cmd := &cobra.Command{
		Use:   "download PIPELINE_UUID [PROJECT/REPO]",
		Short: "Download a pipeline step artifact",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.PipelineUUID = args[0]
			opts.Args = args[1:]
			if runF != nil {
				return runF(opts)
			}
			return downloadRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.StepUUID, "step", "", "Step UUID (required)")
	_ = cmd.MarkFlagRequired("step")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Artifact name (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&opts.Out, "out", "", `Output path ("-" for stdout, default: artifact name in current directory)`)
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func downloadRun(f *factory.Factory, opts *DownloadOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	ac, err := backend.AsPipelineArtifactClient(client, ref.Host)
	if err != nil {
		return err
	}

	var out io.Writer
	var dest string

	switch opts.Out {
	case "-":
		out = f.IOStreams.Out
		dest = ""
	case "":
		dest = opts.Name
		fh, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("create %s: %w", dest, err)
		}
		defer fh.Close() //nolint:errcheck
		out = fh
	default:
		dest = opts.Out
		fh, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("create %s: %w", dest, err)
		}
		defer fh.Close() //nolint:errcheck
		out = fh
	}

	// Use a counting writer to report bytes downloaded.
	cw := &countingWriter{w: out}
	if err := ac.DownloadPipelineArtifact(ref.Project, ref.Slug, opts.PipelineUUID, opts.StepUUID, opts.Name, cw); err != nil {
		return err
	}

	if opts.Out != "-" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Downloaded %s (%d bytes)\n", dest, cw.n)
	}
	return nil
}

// countingWriter wraps an io.Writer and tracks bytes written.
type countingWriter struct {
	w io.Writer
	n int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n)
	return n, err
}
