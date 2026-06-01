package hook

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdHookList builds `repo hook list [PROJECT/REPO] [--json] [--hostname H]`.
func NewCmdHookList(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List plugin hook scripts on a repository",
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

			hc, err := backend.AsRepoHookClient(client, ref.Host)
			if err != nil {
				return err
			}

			hooks, err := hc.ListRepoHooks(ref.Project, ref.Slug)
			if err != nil {
				return err
			}

			p := repoHookFields(f, format.ConfigFromCmd(cmd))
			for _, h := range hooks {
				p.AddItem(h)
			}
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func repoHookFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.RepoHook] {
	p := format.New[backend.RepoHook](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.RepoHook]{
		Name:    "key",
		Header:  "KEY",
		Extract: func(h backend.RepoHook) any { return h.Key },
	})
	p.AddField(format.Field[backend.RepoHook]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(h backend.RepoHook) any { return h.Name },
	})
	p.AddField(format.Field[backend.RepoHook]{
		Name:    "enabled",
		Header:  "ENABLED",
		Extract: func(h backend.RepoHook) any { return h.Enabled },
	})
	p.AddField(format.Field[backend.RepoHook]{
		Name:    "configured",
		Header:  "CONFIGURED",
		Extract: func(h backend.RepoHook) any { return h.Configured },
	})
	return p
}
