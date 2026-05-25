package mirror

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// RepoListOptions carries parsed flags for `mirror repo list`.
type RepoListOptions struct {
	Output   format.OutputConfig
	Hostname string
	Limit    int
	MirrorID string
}

// NewCmdMirrorRepoList constructs the `mirror repo list` cobra command.
func NewCmdMirrorRepoList(f *factory.Factory, runF func(*RepoListOptions) error) *cobra.Command {
	opts := &RepoListOptions{}
	cmd := &cobra.Command{
		Use:   "list <MIRROR_ID>",
		Short: "List repos mirrored by a Smart Mirror server (Server/DC)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			opts.MirrorID = args[0]
			if runF != nil {
				return runF(opts)
			}
			return mirrorRepoListRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of repos (0 = no cap)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (Server/DC only)")
	return cmd
}

func mirrorRepoListRun(f *factory.Factory, opts *RepoListOptions) error {
	host, err := factory.ResolveHost(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	mc, err := backend.AsMirrorClient(client, host)
	if err != nil {
		return err
	}
	repos, listErr := mc.ListMirroredRepos(opts.MirrorID, opts.Limit)
	if listErr != nil && len(repos) == 0 {
		return listErr
	}

	p := mirroredRepoFields(f, opts.Output)
	for _, r := range repos {
		p.AddItem(r)
	}
	if err := p.Render(); err != nil {
		return err
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(repos), listErr)
	return listErr
}

func mirroredRepoFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.MirroredRepo] {
	p := format.New[backend.MirroredRepo](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.MirroredRepo]{
		Name:    "slug",
		Header:  "SLUG",
		Extract: func(r backend.MirroredRepo) any { return r.Slug },
	})
	p.AddField(format.Field[backend.MirroredRepo]{
		Name:    "status",
		Header:  "STATUS",
		Extract: func(r backend.MirroredRepo) any { return r.Status },
	})
	p.AddField(format.Field[backend.MirroredRepo]{
		Name:    "last_sync_at",
		Header:  "LAST SYNC",
		Extract: func(r backend.MirroredRepo) any { return r.LastSyncAt },
	})
	return p
}
