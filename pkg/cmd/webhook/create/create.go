package create

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/webhook/shared"
)

// Options holds parsed flags for `webhook create`.
type Options struct {
	Hostname string
	URL      string
	Events   string // comma-separated
	Secret   string // "-" reads from stdin; @PATH reads from file
	Active   bool

	// Args[0] = PROJECT/REPO
	Args []string

	// Stdin is overridable in tests; defaults to os.Stdin when nil.
	Stdin io.Reader
}

// NewCmdCreate builds the `webhook create` cobra command.
func NewCmdCreate(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{Active: true}
	cmd := &cobra.Command{
		Use:   "create [PROJECT/REPO]",
		Short: "Create a webhook",
		Long: `Create a repository webhook subscribing to one or more events. Pass --url
and --events (comma-separated). The optional --secret enables HMAC signing of
delivery payloads. Use --secret=- to read it from stdin (keeps the value out
of shell history) or --secret=@PATH to read from a file.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return createRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.URL, "url", "", "Webhook delivery URL (required)")
	cmd.Flags().StringVar(&opts.Events, "events", "", "Comma-separated list of event keys (required)")
	cmd.Flags().StringVar(&opts.Secret, "secret", "", "Shared secret (use \"-\" for stdin, \"@PATH\" for file)")
	cmd.Flags().BoolVar(&opts.Active, "active", true, "Whether the webhook is active on creation")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	_ = cmd.MarkFlagRequired("url")
	_ = cmd.MarkFlagRequired("events")
	return cmd
}

func createRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	events := shared.ParseEvents(opts.Events)
	if len(events) == 0 {
		return fmt.Errorf("--events must contain at least one event key")
	}
	secret, err := resolveSecret(opts)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	hook, err := client.CreateWebhook(ref.Project, ref.Slug, backend.CreateWebhookInput{
		URL:    opts.URL,
		Events: events,
		Active: opts.Active,
		Secret: secret,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Created webhook %s -> %s\n", hook.ID, hook.URL)
	return nil
}

// resolveSecret reads the shared secret from stdin (--secret=-), a file
// (--secret=@PATH), or the flag value directly. A trailing newline from stdin
// or file is trimmed since both common shells and editors append one.
func resolveSecret(opts *Options) (string, error) {
	switch {
	case opts.Secret == "":
		return "", nil
	case opts.Secret == "-":
		stdin := opts.Stdin
		if stdin == nil {
			stdin = os.Stdin
		}
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read secret from stdin: %w", err)
		}
		return strings.TrimRight(string(b), "\n"), nil
	case strings.HasPrefix(opts.Secret, "@"):
		b, err := os.ReadFile(opts.Secret[1:])
		if err != nil {
			return "", fmt.Errorf("read secret from %s: %w", opts.Secret[1:], err)
		}
		return strings.TrimRight(string(b), "\n"), nil
	default:
		return opts.Secret, nil
	}
}
