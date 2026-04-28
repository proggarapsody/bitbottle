// Package config implements `bitbottle config` — get/set/list user preferences
// stored in ~/.config/bitbottle/config.yml. Distinct from auth state in
// hosts.yml.
package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdConfig builds the `bitbottle config` parent command.
func NewCmdConfig(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage bitbottle user preferences",
	}
	cmd.AddCommand(newGetCmd(f))
	cmd.AddCommand(newSetCmd(f))
	cmd.AddCommand(newListCmd(f))
	return cmd
}

func newGetCmd(f *factory.Factory) *cobra.Command {
	var host string
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Print the value of a configuration key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.UserConfig()
			if err != nil {
				return err
			}
			v, ok := cfg.Get(args[0], host)
			if !ok {
				return fmt.Errorf("no value set for %q", args[0])
			}
			fmt.Fprintln(f.IOStreams.Out, v)
			return nil
		},
	}
	cmd.Flags().StringVar(&host, "host", "", "Look up the per-host override for this hostname")
	return cmd
}

func newSetCmd(f *factory.Factory) *cobra.Command {
	var host string
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Write a value to the configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.UserConfig()
			if err != nil {
				return err
			}
			if err := cfg.Set(args[0], args[1], host); err != nil {
				return err
			}
			return cfg.Save()
		},
	}
	cmd.Flags().StringVar(&host, "host", "", "Scope this value to a single hostname (per-host keys only)")
	return cmd
}

func newListCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every set configuration value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := f.UserConfig()
			if err != nil {
				return err
			}
			for _, e := range cfg.List() {
				if e.Host == "" {
					fmt.Fprintf(f.IOStreams.Out, "%s=%s\n", e.Key, e.Value)
				} else {
					fmt.Fprintf(f.IOStreams.Out, "%s.%s=%s\n", e.Host, e.Key, e.Value)
				}
			}
			return nil
		},
	}
}
