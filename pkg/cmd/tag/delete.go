package tag

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdTagDelete(f *factory.Factory) *cobra.Command {
	var confirm bool
	var hostname string

	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO NAME",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := f.ResolveRef(args[0], hostname)
			if err != nil {
				return err
			}
			tagName := args[1]

			if !confirm {
				if !f.IOStreams.IsStdoutTTY() {
					return fmt.Errorf("--confirm required when not running interactively")
				}
				fmt.Fprintf(f.IOStreams.Out, "Delete tag %s? [y/N]: ", tagName)
				reader := bufio.NewReader(f.IOStreams.In)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(answer)
				if answer != "y" && answer != "Y" {
					fmt.Fprintln(f.IOStreams.Out, "Aborted.")
					return nil
				}
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			if err := client.DeleteTag(ref.Project, ref.Slug, tagName); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Deleted tag %s\n", tagName)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
