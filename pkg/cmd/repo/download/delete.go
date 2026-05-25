package download

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdDownloadDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete [WORKSPACE/REPO] NAME",
		Short: "Delete a repository download artifact",
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
			if !confirm {
				if !f.IOStreams.IsStdoutTTY() {
					return fmt.Errorf("requires --confirm to delete a download artifact")
				}
				fmt.Fprintf(f.IOStreams.ErrOut, "Delete download %s? [y/N] ", nameArg)
				scanner := bufio.NewScanner(f.IOStreams.In)
				var answer string
				if scanner.Scan() {
					answer = strings.TrimSpace(scanner.Text())
				}
				switch strings.ToLower(answer) {
				case "y", "yes":
					// proceed
				default:
					fmt.Fprintln(f.IOStreams.Out, "Deletion aborted.")
					return nil
				}
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			rc, err := backend.AsRepoDownloadClient(client, ref.Host)
			if err != nil {
				return err
			}
			if err := rc.DeleteRepoDownload(ref.Project, ref.Slug, nameArg); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Deleted %s.\n", nameArg)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
