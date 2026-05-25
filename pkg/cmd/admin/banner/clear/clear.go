// Package clear implements `admin banner clear`.
package clear

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `admin banner clear`.
type Options struct {
	Hostname string
	Confirm  bool
}

// NewCmdBannerClear builds the `admin banner clear` cobra command.
func NewCmdBannerClear(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Remove the site-wide announcement banner",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return bannerClearRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt (required in non-TTY / CI mode)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func bannerClearRun(f *factory.Factory, opts *Options) error {
	if !opts.Confirm {
		if !f.IOStreams.IsStdoutTTY() {
			return fmt.Errorf("--confirm required in non-interactive mode")
		}
		fmt.Fprintln(f.IOStreams.Out, "Remove the site-wide announcement banner? [y/N]")
		reader := bufio.NewReader(f.IOStreams.In)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer != "y" && answer != "Y" {
			fmt.Fprintln(f.IOStreams.Out, "Aborted.")
			return nil
		}
	}

	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	ac, err := backend.AsAdminClient(client, host)
	if err != nil {
		return err
	}
	if err := ac.ClearBanner(); err != nil {
		return err
	}
	fmt.Fprintln(f.IOStreams.Out, "Banner cleared.")
	return nil
}
