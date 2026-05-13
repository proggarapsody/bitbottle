package repo

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdRepoWatcher is the parent of `repo watcher <action>`.
func NewCmdRepoWatcher(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watcher",
		Short: "Manage repository watchers",
	}
	cmd.AddCommand(NewCmdRepoWatcherList(f))
	return cmd
}

// NewCmdRepoWatcherList builds `repo watcher list [PROJECT/REPO]`.
func NewCmdRepoWatcherList(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List users watching a repository",
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

			rw, err := backend.AsRepoWatcherClient(client, ref.Host)
			if err != nil {
				return err
			}

			watchers, err := rw.ListRepoWatchers(ref.Project, ref.Slug)
			if err != nil {
				return err
			}

			p := repoWatcherFields(f, format.ConfigFromCmd(cmd))
			for _, w := range watchers {
				p.AddItem(w)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func repoWatcherFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.User] {
	p := format.New[backend.User](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.User]{
		Name:    "displayName",
		Header:  "DISPLAY_NAME",
		Extract: func(u backend.User) any { return u.DisplayName },
	})
	p.AddField(format.Field[backend.User]{
		Name:    "username",
		Header:  "USERNAME",
		Extract: func(u backend.User) any { return u.Slug },
	})
	return p
}
