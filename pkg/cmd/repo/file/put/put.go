// Package put implements `repo file put`.
package put

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `repo file put`.
type Options struct {
	Hostname     string
	Branch       string
	Message      string
	Content      string
	ContentFile  string
	SourceCommit string
	runF         func(*Options) error
}

// NewCmdFilePut builds `repo file put PATH [PROJECT/REPO] --branch BRANCH --message MSG`.
// Content is supplied via --content TEXT or --content-file FILE (mutually exclusive).
// If PROJECT/REPO is omitted the pinned default repo is used.
func NewCmdFilePut(f *factory.Factory) *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "put PATH [PROJECT/REPO]",
		Short: "Create or update a file on a branch",
		Long: "Write the given content to PATH on BRANCH, creating a commit with MESSAGE.\n\n" +
			"Content is supplied either via --content (inline) or --content-file (from a\n" +
			"local file). Exactly one of these flags is required.\n\n" +
			"If PROJECT/REPO is omitted the repository pinned with `repo set-default` is used.\n\n" +
			"Examples:\n" +
			"  bitbottle repo file put README.md MYWS/my-svc --branch main --message 'Update README' --content '# Hello'\n" +
			"  bitbottle repo file put config.yaml MYWS/my-svc --branch feat/x --message 'Add config' --content-file ./config.yaml",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.runF != nil {
				return opts.runF(opts)
			}
			return putRun(f, opts, args)
		},
	}
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "Target branch name (required)")
	cmd.Flags().StringVar(&opts.Message, "message", "", "Commit message (required)")
	cmd.Flags().StringVar(&opts.Content, "content", "", "File content as a string (mutually exclusive with --content-file)")
	cmd.Flags().StringVar(&opts.ContentFile, "content-file", "", "Path to a local file whose content to write (mutually exclusive with --content)")
	cmd.Flags().StringVar(&opts.SourceCommit, "source-commit", "", "Expected HEAD SHA for conflict detection (optional)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func putRun(f *factory.Factory, opts *Options, args []string) error {
	path := args[0]
	if path == "" {
		return fmt.Errorf("PATH is required")
	}
	if opts.Branch == "" {
		return fmt.Errorf("--branch is required")
	}
	if opts.Message == "" {
		return fmt.Errorf("--message is required")
	}
	if opts.Content != "" && opts.ContentFile != "" {
		return fmt.Errorf("--content and --content-file are mutually exclusive")
	}
	if opts.Content == "" && opts.ContentFile == "" {
		return fmt.Errorf("one of --content or --content-file is required")
	}

	content := opts.Content
	if opts.ContentFile != "" {
		b, err := os.ReadFile(opts.ContentFile)
		if err != nil {
			return fmt.Errorf("reading --content-file: %w", err)
		}
		content = string(b)
	}

	var repoArg string
	if len(args) == 2 {
		repoArg = args[1]
	}

	host, err := resolveHostname(f, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(host)
	if err != nil {
		return err
	}
	sw, err := backend.AsSourceWriter(client, host)
	if err != nil {
		return err
	}

	var ns, slug string
	if repoArg != "" {
		ref, err := bbrepo.Parse(repoArg)
		if err != nil {
			return err
		}
		ns, slug = ref.Project, ref.Slug
	} else {
		base, err := f.BaseRepo()
		if err != nil {
			return err
		}
		ns, slug = base.Project, base.Slug
	}

	if err := sw.PutFile(ns, slug, path, backend.PutFileInput{
		Content:      content,
		Branch:       opts.Branch,
		Message:      opts.Message,
		SourceCommit: opts.SourceCommit,
	}); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Committed %q to %s/%s on branch %s\n", path, ns, slug, opts.Branch)
	return nil
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
