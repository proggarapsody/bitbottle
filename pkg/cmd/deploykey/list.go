package deploykey

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// NewCmdList builds the `deploy-key list` cobra command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	var hostname string
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List deploy keys for a repository",
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
			keys, listErr := dk.ListDeployKeys(ref.Project, ref.Slug)
			if listErr != nil && len(keys) == 0 {
				return listErr
			}
			p := deployKeyFields(f, format.ConfigFromCmd(cmd))
			for _, k := range keys {
				p.AddItem(k)
			}
			if err := p.Render(); err != nil {
				return err
			}
			cmdutil.PartialWarn(f.IOStreams.ErrOut, len(keys), listErr)
			return listErr
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

func deployKeyFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.DeployKey] {
	p := format.New[backend.DeployKey](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.DeployKey]{Name: "id", Header: "ID", Extract: func(k backend.DeployKey) any { return k.ID }})
	p.AddField(format.Field[backend.DeployKey]{Name: "label", Header: "LABEL", Extract: func(k backend.DeployKey) any { return k.Label }})
	p.AddField(format.Field[backend.DeployKey]{Name: "key", Header: "KEY", Extract: func(k backend.DeployKey) any {
		key := k.Key
		if len(key) > 40 {
			key = key[:40] + "..."
		}
		return key
	}})
	p.AddField(format.Field[backend.DeployKey]{Name: "readOnly", Header: "READ-ONLY", Extract: func(k backend.DeployKey) any { return k.ReadOnly }})
	return p
}
