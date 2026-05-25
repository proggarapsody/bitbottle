package version

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdVersionDelete(f *factory.Factory) *cobra.Command {
	var hostname string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete ID [WORKSPACE/REPO]",
		Short: "Delete an issue version from a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			idArg, repoArgs := args[0], args[1:]
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid version ID %q: must be a number", idArg)
			}
			if !confirm {
				if !f.IOStreams.IsStdoutTTY() {
					return fmt.Errorf("--confirm required when not running interactively")
				}
				fmt.Fprintf(f.IOStreams.Out, "Delete version %d? [y/N]: ", id)
				reader := bufio.NewReader(f.IOStreams.In)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(f.IOStreams.Out, "Aborted.")
					return nil
				}
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			vc, err := backend.AsIssueVersionClient(client, ref.Host)
			if err != nil {
				return err
			}
			if err := vc.DeleteIssueVersion(ref.Project, ref.Slug, id); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Deleted version %d.\n", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
