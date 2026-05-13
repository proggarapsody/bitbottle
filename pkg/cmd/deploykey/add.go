package deploykey

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdAdd builds the `deploy-key add` cobra command.
func NewCmdAdd(f *factory.Factory) *cobra.Command {
	var key, label, hostname string
	cmd := &cobra.Command{
		Use:   "add [PROJECT/REPO]",
		Short: "Add a deploy key to a repository",
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
			dk, err := backend.AsDeployKeyClient(client, ref.Host)
			if err != nil {
				return err
			}
			added, err := dk.AddDeployKey(ref.Project, ref.Slug, backend.DeployKeyInput{
				Key:   key,
				Label: label,
			})
			if err != nil {
				return err
			}
			p := deployKeyFields(f, format.ConfigFromCmd(cmd))
			p.AddItem(added)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "SSH public key (required)")
	cmd.Flags().StringVar(&label, "label", "", "Label for the key")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}
