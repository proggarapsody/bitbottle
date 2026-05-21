package edit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdEdit builds the `repo edit` command.
func NewCmdEdit(f *factory.Factory) *cobra.Command {
	var (
		hostname      string
		description   string
		website       string
		language      string
		forkPolicy    string
		enableIssues  bool
		disableIssues bool
		enableWiki    bool
		disableWiki   bool
	)

	cmd := &cobra.Command{
		Use:   "edit [PROJECT/REPO]",
		Short: "Update repository metadata fields",
		Long: `Update mutable repository metadata fields.

On Bitbucket Server / Data Center only --description is forwarded; all
other flags are accepted but silently ignored on Server hosts.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := format.ConfigFromCmd(cmd)

			// Validate mutually exclusive flags before any I/O.
			if enableIssues && disableIssues {
				return fmt.Errorf("--enable-issues and --disable-issues are mutually exclusive")
			}
			if enableWiki && disableWiki {
				return fmt.Errorf("--enable-wiki and --disable-wiki are mutually exclusive")
			}

			// Build input — check at least one flag was provided.
			in := backend.EditRepoInput{}
			noop := true

			if cmd.Flags().Changed("description") {
				in.Description = &description
				noop = false
			}
			if cmd.Flags().Changed("website") {
				in.Website = &website
				noop = false
			}
			if cmd.Flags().Changed("language") {
				in.Language = &language
				noop = false
			}
			if cmd.Flags().Changed("fork-policy") {
				in.ForkPolicy = &forkPolicy
				noop = false
			}
			if enableIssues {
				v := true
				in.HasIssues = &v
				noop = false
			}
			if disableIssues {
				v := false
				in.HasIssues = &v
				noop = false
			}
			if enableWiki {
				v := true
				in.HasWiki = &v
				noop = false
			}
			if disableWiki {
				v := false
				in.HasWiki = &v
				noop = false
			}

			if noop {
				return fmt.Errorf("at least one flag is required (--description, --website, --language, --fork-policy, --enable-issues, --disable-issues, --enable-wiki, --disable-wiki)")
			}

			// Resolve repository reference.
			var ns, slug string
			if len(args) == 1 {
				ref, err := bbrepo.Parse(args[0])
				if err != nil {
					return err
				}
				ns = ref.Project
				slug = ref.Slug
			} else {
				ref, err := f.BaseRepo()
				if err != nil {
					return err
				}
				ns = ref.Project
				slug = ref.Slug
			}

			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(host)
			if err != nil {
				return err
			}

			editor, err := backend.AsRepoEditor(client, host)
			if err != nil {
				return err
			}

			repo, err := editor.EditRepo(ns, slug, in)
			if err != nil {
				return err
			}

			if cfg.Format != format.FormatTable {
				p := repoFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(repo)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ Repository %s/%s updated.\n", ns, slug)
			return nil
		},
	}

	format.RegisterOutputFlags(cmd)
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&description, "description", "", "Update repository description")
	cmd.Flags().StringVar(&website, "website", "", "Update repository website URL (Cloud only)")
	cmd.Flags().StringVar(&language, "language", "", "Update repository programming language (Cloud only)")
	cmd.Flags().StringVar(&forkPolicy, "fork-policy", "", "Set fork policy: allow_forks|no_public_forks|no_forks (Cloud only)")
	cmd.Flags().BoolVar(&enableIssues, "enable-issues", false, "Enable issue tracker (Cloud only)")
	cmd.Flags().BoolVar(&disableIssues, "disable-issues", false, "Disable issue tracker (Cloud only)")
	cmd.Flags().BoolVar(&enableWiki, "enable-wiki", false, "Enable wiki (Cloud only)")
	cmd.Flags().BoolVar(&disableWiki, "disable-wiki", false, "Disable wiki (Cloud only)")

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
