// Package delete implements `profile delete NAME`.
package delete

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `profile delete`.
type Options struct {
	Name    string // Args[0]
	Confirm bool   // --confirm
}

// NewCmdDelete builds the `profile delete NAME` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a named credential profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	return cmd
}

func deleteRun(f *factory.Factory, opts *Options) error {
	store, err := f.Profiles()
	if err != nil {
		return err
	}
	if _, ok := store.Get(opts.Name); !ok {
		return fmt.Errorf("profile %q not found", opts.Name)
	}

	if !opts.Confirm {
		if !f.IOStreams.IsStdoutTTY() {
			return fmt.Errorf("--confirm required when not running interactively")
		}
		fmt.Fprintf(f.IOStreams.Out, "Delete profile %s? [y/N]: ", opts.Name)
		reader := bufio.NewReader(f.IOStreams.In)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer != "y" && answer != "Y" {
			fmt.Fprintln(f.IOStreams.Out, "Aborted.")
			return nil
		}
	}

	store.Delete(opts.Name)
	if err := store.Save(); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Deleted profile %s\n", opts.Name)
	return nil
}
