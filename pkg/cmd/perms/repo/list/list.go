package list

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/perms/shared"
)

// Options holds parsed flags for `perms repo list`.
type Options struct {
	Hostname string
	Output   format.OutputConfig
	// Args[0] = PROJECT/REPO
	Args []string
}

// NewCmdList builds the `perms repo list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list PROJECT/REPO",
		Short: "List permissions for a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	project, slug, err := parseProjectRepo(opts.Args[0])
	if err != nil {
		return err
	}
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPermissionsClient(client, host)
	if err != nil {
		return err
	}
	grants, err := pc.ListRepoPermissions(context.Background(), project, slug)
	if err != nil {
		return err
	}
	p := shared.GrantPrinter(f, opts.Output)
	for _, g := range grants {
		p.AddItem(g)
	}
	return p.Render()
}

func parseProjectRepo(s string) (project, slug string, err error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid argument %q: expected PROJECT/REPO", s)
	}
	return parts[0], parts[1], nil
}
