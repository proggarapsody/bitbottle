package label

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdLabelUpdate builds `repo label update [PROJECT/REPO] ID [--name N] [--color C]`.
func NewCmdLabelUpdate(f *factory.Factory) *cobra.Command {
	var hostname, name, color string

	cmd := &cobra.Command{
		Use:   "update [PROJECT/REPO] ID",
		Short: "Update a label on a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, idArg := splitLabelIDArg(args)
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid label ID %q: must be a number", idArg)
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			rl, err := backend.AsRepoLabelClient(client, ref.Host)
			if err != nil {
				return err
			}
			lbl, err := rl.UpdateRepoLabel(ref.Project, ref.Slug, id, backend.UpdateRepoLabelInput{
				Name:  name,
				Color: color,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Updated label %q (ID %d)\n", lbl.Name, lbl.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&name, "name", "", "New label name")
	cmd.Flags().StringVar(&color, "color", "", "New label color (hex)")
	return cmd
}

// splitLabelIDArg handles: [PROJECT/REPO] ID
// Returns (repoArgs, labelID).
func splitLabelIDArg(args []string) ([]string, string) {
	if len(args) == 2 {
		return []string{args[0]}, args[1]
	}
	return nil, args[0]
}
