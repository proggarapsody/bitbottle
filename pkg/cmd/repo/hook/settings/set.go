package settings

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/repoarg"
)

// NewCmdSettingsSet builds `repo hook settings set [PROJECT/REPO] HOOK_KEY --config-file FILE [--hostname H]`.
func NewCmdSettingsSet(f *factory.Factory) *cobra.Command {
	var hostname, configFile string

	cmd := &cobra.Command{
		Use:   "set [PROJECT/REPO] HOOK_KEY",
		Short: "Set JSON settings for a plugin hook script",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if configFile == "" {
				return fmt.Errorf("--config-file is required")
			}

			repoArgs, rest := repoarg.SplitLeadingRepo(args, 1)
			hookKey := rest[0]

			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}

			// Read config JSON from file or stdin.
			var raw []byte
			if configFile == "-" {
				raw, err = io.ReadAll(f.IOStreams.In)
			} else {
				raw, err = os.ReadFile(configFile)
			}
			if err != nil {
				return fmt.Errorf("reading config file: %w", err)
			}

			// Validate it's valid JSON before sending.
			if !json.Valid(raw) {
				return fmt.Errorf("config file does not contain valid JSON")
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			hc, err := backend.AsRepoHookClient(client, ref.Host)
			if err != nil {
				return err
			}

			if err := hc.SetRepoHookSettings(ref.Project, ref.Slug, hookKey, json.RawMessage(raw)); err != nil {
				return err
			}

			fmt.Fprintf(f.IOStreams.Out, "Updated settings for hook %s\n", hookKey)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&configFile, "config-file", "", `JSON config file path, or "-" for stdin (required)`)
	return cmd
}
