package delete

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `webhook delete`.
type Options struct {
	Hostname string
	Confirm  bool

	// Args[0] = PROJECT/REPO, Args[1] = ID
	Args []string
}

// NewCmdDelete builds the `webhook delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO ID",
		Short: "Delete a webhook",
		Args:  cobra.ExactArgs(2),
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
	ref, err := factory.ResolveTarget(f, opts.Args[:1], opts.Hostname)
	if err != nil {
		return err
	}
	id := opts.Args[1]

	if !opts.Confirm {
		if !f.IOStreams.IsStdoutTTY() {
			return fmt.Errorf("--confirm required when not running interactively")
		}
		fmt.Fprintf(f.IOStreams.Out, "Delete webhook %s? [y/N]: ", id)
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
	if err := client.DeleteWebhook(ref.Project, ref.Slug, id); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Deleted webhook %s\n", id)
	return nil
}
