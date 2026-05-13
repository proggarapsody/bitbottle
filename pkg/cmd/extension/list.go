package extension

import (
	"fmt"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/extensions"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdList returns the `extension list` subcommand.
func NewCmdList(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed extensions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			extDir := filepath.Join(f.ConfigDir(), "extensions")
			mgr := extensions.New(extDir, nil)

			exts, err := mgr.List()
			if err != nil {
				return err
			}

			if len(exts) == 0 {
				fmt.Fprint(f.IOStreams.Out, "No extensions installed.\n")
				return nil
			}

			tw := tabwriter.NewWriter(f.IOStreams.Out, 0, 8, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tVERSION\tSOURCE")
			for _, e := range exts {
				version := e.Version
				source := e.Repo
				if e.Local {
					version = "(local)"
					source = e.LocalPath
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Name, version, source)
			}
			return tw.Flush()
		},
	}
}
