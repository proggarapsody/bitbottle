package settings

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// NewCmdSettingsGet builds `repo hook settings get [PROJECT/REPO] HOOK_KEY [--hostname H]`.
func NewCmdSettingsGet(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "get [PROJECT/REPO] HOOK_KEY",
		Short: "Get raw JSON settings for a plugin hook script",
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

			raw, err := hc.GetRepoHookSettings(ref.Project, ref.Slug, hookKey)
			if err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "%s\n", raw)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
