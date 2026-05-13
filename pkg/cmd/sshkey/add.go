package sshkey

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdAdd builds the `ssh-key add` cobra command.
func NewCmdAdd(f *factory.Factory) *cobra.Command {
	var key, label, hostname string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an SSH key for the current user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := factory.ResolveHost(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			sk, err := backend.AsSSHKeyClient(client, host)
			if err != nil {
				return err
			}
			added, err := sk.AddSSHKey(backend.SSHKeyInput{
				Key:   key,
				Label: label,
			})
			if err != nil {
				return err
			}
			p := sshKeyFields(f, format.ConfigFromCmd(cmd))
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
