// Package alias implements `bitbottle alias` — user-defined command shortcuts.
package alias

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/aliases"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdAlias builds the `bitbottle alias` parent. builtinNames is the list of
// top-level commands the user must not shadow; main wires in the live list
// from root.
func NewCmdAlias(f *factory.Factory, builtinNames []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias <command>",
		Short: "Manage command aliases",
	}
	cmd.AddCommand(newSetCmd(f, builtinNames))
	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newDeleteCmd(f))
	return cmd
}

func newSetCmd(f *factory.Factory, builtinNames []string) *cobra.Command {
	return &cobra.Command{
		Use:   "set <name> <expansion>",
		Short: "Create or update an alias (prefix expansion with `!` for shell aliases)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := aliases.CheckShadow(args[0], builtinNames); err != nil {
				return err
			}
			store, err := f.Aliases()
			if err != nil {
				return err
			}
			if err := store.Set(args[0], args[1]); err != nil {
				return err
			}
			return store.Save()
		},
	}
}

func newListCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all aliases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := f.Aliases()
			if err != nil {
				return err
			}
			for _, e := range store.List() {
				fmt.Fprintf(f.IOStreams.Out, "%s: %s\n", e.Name, e.Expansion)
			}
			return nil
		},
	}
}

func newDeleteCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <name>",
		Aliases: []string{"remove", "rm"},
		Short:   "Remove an alias",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := f.Aliases()
			if err != nil {
				return err
			}
			if !store.Delete(args[0]) {
				return fmt.Errorf("no alias named %q", args[0])
			}
			return store.Save()
		},
	}
}
