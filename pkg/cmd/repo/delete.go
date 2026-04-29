package repo

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoDelete(f *factory.Factory) *cobra.Command {
	var confirm bool
	var hostname string

	cmd := &cobra.Command{
		Use:   "delete [PROJECT/REPO]",
		Short: "Delete a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			if !confirm {
				proceed, err := confirmDelete(f, ref)
				if err != nil {
					return err
				}
				if !proceed {
					fmt.Fprintln(f.IOStreams.Out, "Deletion aborted.")
					return nil
				}
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			if err := client.DeleteRepo(ref.Project, ref.Slug); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Deleted repository %s/%s\n", ref.Project, ref.Slug)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

// confirmDelete prompts the user interactively for delete confirmation.
// Returns (false, error) when running non-TTY (which requires --confirm),
// (true, nil) when the user answered "y"/"yes", and (false, nil) on any
// other answer.
func confirmDelete(f *factory.Factory, ref bbrepo.RepoRef) (bool, error) {
	if !f.IOStreams.IsStdoutTTY() {
		return false, fmt.Errorf("requires --confirm to delete a repository")
	}
	fmt.Fprintf(f.IOStreams.ErrOut, "Are you sure you want to delete %s/%s? [y/N] ",
		ref.Project, ref.Slug)

	scanner := bufio.NewScanner(f.IOStreams.In)
	var answer string
	if scanner.Scan() {
		answer = strings.TrimSpace(scanner.Text())
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
