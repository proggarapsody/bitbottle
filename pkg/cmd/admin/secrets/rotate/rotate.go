// Package rotate implements `admin secrets rotate`.
package rotate

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

const sysAdminHint = "This requires SYS_ADMIN permission. Standard admin tokens do not include it; the action must be performed by a system administrator."

// Options holds parsed flags for `admin secrets rotate`.
type Options struct {
	Hostname string
	Confirm  bool
}

// NewCmdRotate builds the `admin secrets rotate` cobra command.
func NewCmdRotate(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate the cluster HTTPS secret (DC deployments)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return rotateRun(f, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func rotateRun(f *factory.Factory, opts *Options) error {
	if !opts.Confirm {
		if !f.IOStreams.IsStdoutTTY() {
			return fmt.Errorf("--confirm required in non-interactive mode")
		}
		fmt.Fprintln(f.IOStreams.Out, "This rotates the cluster's internal HTTPS secret. ALL nodes must be restarted for the new secret to take effect. Continue? [y/N]")
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
	if err := ac.RotateSecrets(); err != nil {
		var de *backend.DomainError
		if errors.As(err, &de) && de.Kind == backend.ErrPermission {
			fmt.Fprintln(f.IOStreams.ErrOut, sysAdminHint)
		}
		return err
	}
	fmt.Fprintln(f.IOStreams.Out, "Secrets rotated. Restart all cluster nodes for the new secret to take effect.")
	return nil
}
