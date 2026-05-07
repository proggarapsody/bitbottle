package protect

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// DeleteOptions holds parsed flags for `branch protect delete`.
type DeleteOptions struct {
	Hostname string

	// Args[0] = PROJECT/REPO, Args[1] = ID (numeric)
	Args []string
}

// NewCmdDelete builds the `branch protect delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete PROJECT/REPO ID",
		Short: "Remove a branch restriction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return deleteRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func deleteRun(f *factory.Factory, opts *DeleteOptions) error {
	id, err := strconv.Atoi(opts.Args[1])
	if err != nil {
		return fmt.Errorf("ID must be numeric (got %q): %w", opts.Args[1], err)
	}
	ref, err := factory.ResolveTarget(f, opts.Args[:1], opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	bp, err := backend.AsBranchProtector(client, ref.Host)
	if err != nil {
		return err
	}
	if err := bp.DeleteBranchProtection(ref.Project, ref.Slug, id); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Deleted restriction %d\n", id)
	return nil
}
