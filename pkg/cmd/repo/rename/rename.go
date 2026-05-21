package rename

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdRename(f *factory.Factory) *cobra.Command {
	var hostname string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "rename PROJECT/REPO NEW-NAME",
		Short: "Rename a repository (both backends)",
		Long: "Rename a repository on Bitbucket Cloud or Server / Data Center.\n" +
			"On Cloud the slug is derived from the new name (e.g. \"My New Name\" → \"my-new-name\");\n" +
			"existing clones must update their `origin` URL — `git remote set-url origin ...` —\n" +
			"after a rename. Because of that, --confirm is required when not running on a TTY.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}
			newName := args[1]

			if !confirm {
				proceed, err := confirmRename(f, ref, newName)
				if err != nil {
					return err
				}
				if !proceed {
					fmt.Fprintln(f.IOStreams.Out, "Rename aborted.")
					return nil
				}
			}

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			updated, err := client.RenameRepo(ref.Project, ref.Slug, newName)
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := repoFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(updated)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "Renamed %s/%s to %s/%s\n",
				ref.Project, ref.Slug, updated.Namespace, updated.Slug)
			return nil
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func confirmRename(f *factory.Factory, ref bbrepo.RepoRef, newName string) (bool, error) {
	if !f.IOStreams.IsStdoutTTY() {
		return false, fmt.Errorf("requires --confirm to rename a repository (slug change breaks existing clones' origin URL)")
	}
	fmt.Fprintf(f.IOStreams.ErrOut, "Rename %s/%s to %q? Existing clones' `origin` URL will need updating. [y/N] ",
		ref.Project, ref.Slug, newName)

	scanner := bufio.NewScanner(f.IOStreams.In)
	var answer string
	if scanner.Scan() {
		answer = strings.TrimSpace(scanner.Text())
	}
	switch strings.ToLower(answer) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
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
