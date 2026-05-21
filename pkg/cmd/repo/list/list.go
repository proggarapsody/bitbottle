package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

func NewCmdList(f *factory.Factory) *cobra.Command {
	var limit int
	var hostname string

	cmd := &cobra.Command{
		Use:   "list [WORKSPACE]",
		Short: "List repositories",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(limit); err != nil {
				return err
			}
			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			ns := ""
			if len(args) > 0 {
				ns = args[0]
			}

			repos, err := client.ListRepos(ns, limit)
			if err != nil {
				return err
			}

			p := repoFields(f, format.ConfigFromCmd(cmd))
			for _, r := range repos {
				p.AddItem(r)
			}
			return p.Render()
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 30, "Maximum number of repositories")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (defaults to configured host)")
	return cmd
}

func repoFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Repository] {
	p := format.New[backend.Repository](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Repository]{Name: "slug", Header: "SLUG", Extract: func(r backend.Repository) any { return r.Slug }})
	p.AddField(format.Field[backend.Repository]{Name: "name", Header: "NAME", Extract: func(r backend.Repository) any { return r.Name }})
	p.AddField(format.Field[backend.Repository]{Name: "namespace", Header: "PROJECT", Extract: func(r backend.Repository) any { return r.Namespace }})
	p.AddField(format.Field[backend.Repository]{Name: "scm", Header: "TYPE", Extract: func(r backend.Repository) any { return r.SCM }})
	p.AddField(format.Field[backend.Repository]{Name: "description", Header: "DESCRIPTION", Extract: func(r backend.Repository) any { return r.Description }})
	p.AddField(format.Field[backend.Repository]{Name: "webURL", Header: "URL", Extract: func(r backend.Repository) any { return r.WebURL }})
	return p
}

func resolveHostname(f *factory.Factory, flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	hosts := cfg.Hosts()
	switch len(hosts) {
	case 0:
		return "", fmt.Errorf("not authenticated; run `bitbottle auth login` first")
	case 1:
		return hosts[0], nil
	default:
		return "", fmt.Errorf("multiple hosts configured; use --hostname to specify one")
	}
}
