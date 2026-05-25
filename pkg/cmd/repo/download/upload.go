package download

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdDownloadUpload(f *factory.Factory) *cobra.Command {
	var hostname, name string

	cmd := &cobra.Command{
		Use:   "upload [WORKSPACE/REPO] FILE",
		Short: "Upload a file as a repository download artifact",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var repoArgs []string
			var fileArg string
			if len(args) == 2 {
				repoArgs = []string{args[0]}
				fileArg = args[1]
			} else {
				fileArg = args[0]
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
			uploadName := name
			if uploadName == "" {
				uploadName = filepath.Base(fileArg)
			}
			fh, err := os.Open(fileArg)
			if err != nil {
				return fmt.Errorf("opening file %s: %w", fileArg, err)
			}
			defer fh.Close()
			if _, err := rc.UploadRepoDownload(ref.Project, ref.Slug, uploadName, fh); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Uploaded %s.\n", uploadName)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Name to use for the download artifact (default: filename)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
