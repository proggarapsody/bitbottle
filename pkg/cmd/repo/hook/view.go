package hook

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// NewCmdHookView builds `repo hook view [PROJECT/REPO] HOOK_KEY [--json] [--hostname H]`.
func NewCmdHookView(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "view [PROJECT/REPO] HOOK_KEY",
		Short: "Show details for a single plugin hook script",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArgs, rest := repoarg.SplitLeadingRepo(args, 1)
			hookKey := rest[0]

			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
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

			h, err := hc.GetRepoHook(ref.Project, ref.Slug, hookKey)
			if err != nil {
				return err
			}

			p := repoHookViewFields(f, format.ConfigFromCmd(cmd))
			p.SetSingleItem()
			p.AddItem(h)
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func repoHookViewFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.RepoHook] {
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
		Name:    "version",
		Header:  "VERSION",
		Extract: func(h backend.RepoHook) any { return h.Version },
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
