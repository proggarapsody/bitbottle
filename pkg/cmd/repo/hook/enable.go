package hook

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// NewCmdHookEnable builds `repo hook enable [PROJECT/REPO] HOOK_KEY [--hostname H]`.
func NewCmdHookEnable(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "enable [PROJECT/REPO] HOOK_KEY",
		Short: "Enable a plugin hook script on a repository",
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

			if err := hc.EnableRepoHook(ref.Project, ref.Slug, hookKey); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Enabled hook %s\n", hookKey)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}
