package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdVersionCreate(f *factory.Factory) *cobra.Command {
	var hostname string
	var name string

	cmd := &cobra.Command{
		Use:   "create [WORKSPACE/REPO]",
		Short: "Create an issue version in a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
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
			v, err := vc.CreateIssueVersion(ref.Project, ref.Slug, name)
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Created version %q (ID: %d).\n", v.Name, v.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Version name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
