package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCreate(f *factory.Factory) *cobra.Command {
	var hostname, project, description string
	var private bool

	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "Create a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if project == "" {
				return fmt.Errorf("--project is required")
			}

			name := args[0]

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			r, err := client.CreateRepo(project, backend.CreateRepoInput{
				Name:        name,
				SCM:         "git",
				Public:      !private,
				Description: description,
			})
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := repoFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(r)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Created repository %s/%s\n", r.Namespace, r.Slug)
			if r.WebURL != "" {
				fmt.Fprintf(f.IOStreams.Out, "%s\n", r.WebURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&project, "project", "", "Project key")
	cmd.Flags().StringVar(&description, "description", "", "Repository description")
	cmd.Flags().BoolVar(&private, "private", true, "Make repository private")
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
