package delete

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `pipeline variable delete`.
type Options struct {
	Hostname string
	Confirm  bool

	// Args[0] = PROJECT/REPO, Args[1] = KEY
	Args []string
}

// NewCmdDelete builds the `pipeline variable delete` cobra command.
//
// Deprecated: use `bitbottle variable delete --scope repository` instead.
func NewCmdDelete(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:        "delete PROJECT/REPO KEY",
		Short:      "Delete a pipeline variable",
		Deprecated: "use `bitbottle variable delete --scope repository` instead",
		Args:       cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func deleteRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	key := opts.Args[1]

	if !opts.Confirm {
		if !f.IOStreams.IsStdoutTTY() {
			return fmt.Errorf("--confirm required when not running interactively")
		}
		fmt.Fprintf(f.IOStreams.Out, "Delete pipeline variable %s? [y/N]: ", key)
		reader := bufio.NewReader(f.IOStreams.In)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer != "y" && answer != "Y" {
			fmt.Fprintln(f.IOStreams.Out, "Aborted.")
			return nil
		}
	}

	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPipelineClient(client, ref.Host)
	if err != nil {
		return err
	}
	if err := pc.DeletePipelineVariable(ref.Project, ref.Slug, key); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Deleted pipeline variable %s\n", key)
	return nil
}
