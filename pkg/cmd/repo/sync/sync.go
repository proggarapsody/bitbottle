// Package sync implements the "repo sync" command which synchronises a Cloud
// fork branch with its upstream repository.
package sync

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdSync(f *factory.Factory) *cobra.Command {
	var branch, hostname string

	cmd := &cobra.Command{
		Use:   "sync [PROJECT/REPO]",
		Short: "Sync a fork with its upstream repository",
		Long: "Synchronise a Bitbucket Cloud fork branch with its upstream repository.\n" +
			"Bitbucket Server / Data Center has no fork-upstream concept — running this\n" +
			"against a Server host returns a typed unsupported-capability error.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			syncer, err := backend.AsRepoSyncer(client, ref.Host)
			if err != nil {
				return err
			}

			result, err := syncer.SyncRepo(ref.Project, ref.Slug, branch)
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := syncResultFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(result)
				return p.Render()
			}

			if result.CommitsMerged == 0 {
				fmt.Fprintln(f.IOStreams.Out, "Already up to date")
			} else {
				fmt.Fprintf(f.IOStreams.Out, "Synced %d commit(s) from upstream\n", result.CommitsMerged)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&branch, "branch", "", "Branch to sync (default: repo's default branch)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func syncResultFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.SyncResult] {
	p := format.New[backend.SyncResult](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.SyncResult]{
		Name:    "behind",
		Header:  "BEHIND",
		Extract: func(r backend.SyncResult) any { return r.Behind },
	})
	p.AddField(format.Field[backend.SyncResult]{
		Name:    "commits_merged",
		Header:  "COMMITS_MERGED",
		Extract: func(r backend.SyncResult) any { return r.CommitsMerged },
	})
	return p
}
