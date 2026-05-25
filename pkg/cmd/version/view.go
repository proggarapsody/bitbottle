package version

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdVersionView(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "view ID [WORKSPACE/REPO]",
		Short: "View a single issue version",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			idArg, repoArgs := args[0], args[1:]
			id, err := strconv.Atoi(idArg)
			if err != nil {
				return fmt.Errorf("invalid version ID %q: must be a number", idArg)
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
			v, err := vc.GetIssueVersion(ref.Project, ref.Slug, id)
			if err != nil {
				return err
			}
			p := versionListFields(f, format.ConfigFromCmd(cmd))
			p.SetSingleItem()
			p.AddItem(v)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}
