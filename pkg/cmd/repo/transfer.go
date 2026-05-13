package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRepoTransfer(f *factory.Factory) *cobra.Command {
	var hostname, to string

	cmd := &cobra.Command{
		Use:   "transfer [PROJECT/REPO] --to TARGET",
		Short: "Transfer a repository to another project or workspace",
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
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			rt, err := backend.AsRepoTransferClient(client, host)
			if err != nil {
				return err
			}

			updated, err := rt.TransferRepo(ref.Project, ref.Slug, to)
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := repoFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(updated)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Transferred to %s/%s\n", to, updated.Slug)
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&to, "to", "", "Target project key (Server) or workspace slug (Cloud)")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}
