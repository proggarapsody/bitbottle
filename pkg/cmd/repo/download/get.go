package download

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdDownloadGet(f *factory.Factory) *cobra.Command {
	var hostname, outPath string

	cmd := &cobra.Command{
		Use:   "get [WORKSPACE/REPO] NAME",
		Short: "Download a repository download artifact to a local file",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var repoArgs []string
			var nameArg string
			if len(args) == 2 {
				repoArgs = []string{args[0]}
				nameArg = args[1]
			} else {
				nameArg = args[0]
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
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
			dest := outPath
			if dest == "" {
				dest = nameArg
			}
			fh, err := os.Create(dest)
			if err != nil {
				return fmt.Errorf("creating output file %s: %w", dest, err)
			}
			defer fh.Close()
			if err := rc.DownloadRepoDownload(ref.Project, ref.Slug, nameArg, fh); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Downloaded %s to %s.\n", nameArg, dest)
			return nil
		},
	}
	cmd.Flags().StringVar(&outPath, "out", "", "Output file path (default: artifact name in current directory)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
