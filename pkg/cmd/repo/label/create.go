package label

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdLabelCreate builds `repo label create [PROJECT/REPO] --name N [--color C]`.
func NewCmdLabelCreate(f *factory.Factory) *cobra.Command {
	var hostname, name, color string

	cmd := &cobra.Command{
		Use:   "create [PROJECT/REPO]",
		Short: "Create a label on a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			ref, err := factory.ResolveTarget(f, args, hostname)
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
			lbl, err := rl.CreateRepoLabel(ref.Project, ref.Slug, backend.CreateRepoLabelInput{
				Name:  name,
				Color: color,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Created label %q (ID %d)\n", lbl.Name, lbl.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&name, "name", "", "Label name (required)")
	cmd.Flags().StringVar(&color, "color", "", "Label color (hex, e.g. #ff0000)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
