package sshkey

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdList builds the `ssh-key list` cobra command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSH keys for the current user",
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
			keys, err := sk.ListSSHKeys()
			if err != nil {
				return err
			}
			p := sshKeyFields(f, format.ConfigFromCmd(cmd))
			for _, k := range keys {
				p.AddItem(k)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func sshKeyFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.SSHKey] {
	p := format.New[backend.SSHKey](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.SSHKey]{Name: "id", Header: "ID", Extract: func(k backend.SSHKey) any { return k.ID }})
	p.AddField(format.Field[backend.SSHKey]{Name: "label", Header: "LABEL", Extract: func(k backend.SSHKey) any { return k.Label }})
	p.AddField(format.Field[backend.SSHKey]{Name: "key", Header: "KEY", Extract: func(k backend.SSHKey) any {
		key := k.Key
		if len(key) > 40 {
			key = key[:40] + "..."
		}
		return key
	}})
	return p
}
