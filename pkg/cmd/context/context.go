// Package context implements the `bitbottle context` command — a single
// one-call orientation primitive that returns the current host, project,
// slug, branch, default branch, ahead/behind counts vs the default branch,
// authenticated user, and backend type. This collapses what previously
// required three independent calls (auth status / repo view / git status)
// into a single structured response, principally for AI-agent integrations
// that pay a per-round-trip latency cost.
package context

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdContext builds the cobra command. The command has no subcommands and
// takes no positional arguments — the entire shape is resolved from config,
// git state, and one backend round-trip (current user + branch list when
// inside a repo).
func NewCmdContext(f *factory.Factory) *cobra.Command {
	var jsonFields string
	var jqExpr string
	var hostname string

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Show host, repo, branch, user, and backend in one call",
		Long: "Returns a single structured snapshot of where you are: host,\n" +
			"project, slug, current branch, default branch, ahead/behind vs\n" +
			"the default branch, authenticated user, and backend type.\n\n" +
			"Outside a git repository project / slug / branch / default_branch\n" +
			"and ahead / behind are empty / zero, but host, user, and backend\n" +
			"still resolve via the configured (or --hostname) host.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := Build(f, hostname)
			if err != nil {
				return err
			}

			if jsonFields != "" || jqExpr != "" {
				p := contextFields(f, jsonFields, jqExpr)
				p.SetSingleItem()
				p.AddItem(ctx)
				return p.Render()
			}

			return renderTable(f, ctx)
		},
	}
	cmd.Flags().StringVar(&jsonFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output with a jq expression")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

// Build assembles a backend.Context from config, git state, and the backend
// (current user + branch list when inside a repo). Exposed so the MCP
// `get_context` handler returns the exact same shape as the CLI.
//
// Outside a git repository, BaseRepo() returns a non-nil error. That branch
// yields an empty Project / Slug / Branch / DefaultBranch / Ahead / Behind
// while still resolving Host, User, and Backend.
func Build(f *factory.Factory, hostname string) (backend.Context, error) {
	ref, baseRepoErr := f.BaseRepo()

	host := hostname
	if host == "" {
		host = ref.Host
	}
	if host == "" {
		// No -R, no remote, no flag — fall back to the single-host rule.
		resolved, err := factory.ResolveHost(f, "")
		if err != nil {
			return backend.Context{}, err
		}
		host = resolved
	}

	client, err := f.Backend(host)
	if err != nil {
		return backend.Context{}, err
	}

	user, err := client.GetCurrentUser()
	if err != nil {
		return backend.Context{}, err
	}

	ctx := backend.Context{
		Host:    host,
		Backend: backendKind(f, host),
		User:    backend.ContextUser(user),
	}

	if baseRepoErr != nil {
		// Outside a repo: the empty repo-scoped fields are intentional. We
		// still want the user + host + backend so agents can orient.
		return ctx, nil
	}

	ctx.Project = ref.Project
	ctx.Slug = ref.Slug

	g := f.GitRunner()
	if branch, berr := currentBranch(g); berr == nil {
		ctx.Branch = branch
	}

	if def, derr := backend.DefaultBranch(client, ref.Project, ref.Slug); derr == nil {
		ctx.DefaultBranch = def
	}
	if ctx.DefaultBranch != "" && ctx.Branch != "" {
		ahead, behind, _ := aheadBehind(g, ctx.DefaultBranch)
		ctx.Ahead = ahead
		ctx.Behind = behind
	}

	return ctx, nil
}

// backendKind returns "cloud" or "server" — the same vocabulary
// bbinstance.IsCloud uses, exposed in the JSON shape so callers don't have
// to re-derive it from the hostname.
func backendKind(f *factory.Factory, host string) string {
	cfg, err := f.Config()
	var backendType string
	if err == nil && cfg != nil {
		hc, _ := cfg.Get(host)
		backendType = hc.BackendType
	}
	if bbinstance.IsCloud(host, backendType) {
		return bbinstance.BackendTypeCloud
	}
	return bbinstance.BackendTypeServer
}

// currentBranch wraps `git rev-parse --abbrev-ref HEAD`. Returns an error
// (and an empty string) when the runner fails, e.g. detached HEAD or not
// a git repository.
func currentBranch(runner interface {
	Run(args ...string) (string, string, error)
}) (string, error) {
	out, _, err := runner.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	br := strings.TrimSpace(out)
	if br == "" || br == "HEAD" {
		return "", fmt.Errorf("detached or empty HEAD")
	}
	return br, nil
}

// aheadBehind computes ahead/behind counts of HEAD vs base using
// `git rev-list --left-right --count base...HEAD`. Output shape is
// "<behind>\t<ahead>" — left is base-ahead-of-HEAD (i.e. behind), right
// is HEAD-ahead-of-base. We return (ahead, behind) for the caller's
// natural reading order.
//
// Any failure (no upstream, base ref missing, not a git repo) returns
// zeros and a nil error so the caller can degrade gracefully — outside
// a repository or pre-fetch, ahead/behind is genuinely "unknown" and
// the documented shape uses 0 to mean exactly that.
func aheadBehind(runner interface {
	Run(args ...string) (string, string, error)
}, base string) (ahead, behind int, err error) {
	if base == "" {
		return 0, 0, nil
	}
	out, _, runErr := runner.Run("rev-list", "--left-right", "--count", base+"...HEAD")
	if runErr != nil {
		return 0, 0, nil
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		return 0, 0, nil
	}
	left, lerr := strconv.Atoi(parts[0])
	right, rerr := strconv.Atoi(parts[1])
	if lerr != nil || rerr != nil {
		return 0, 0, nil
	}
	return right, left, nil
}

func renderTable(f *factory.Factory, ctx backend.Context) error {
	w := f.IOStreams.Out
	fmt.Fprintf(w, "Host:           %s\n", ctx.Host)
	fmt.Fprintf(w, "Backend:        %s\n", ctx.Backend)
	if ctx.Project != "" || ctx.Slug != "" {
		fmt.Fprintf(w, "Project:        %s\n", ctx.Project)
		fmt.Fprintf(w, "Slug:           %s\n", ctx.Slug)
	}
	if ctx.Branch != "" {
		fmt.Fprintf(w, "Branch:         %s\n", ctx.Branch)
	}
	if ctx.DefaultBranch != "" {
		fmt.Fprintf(w, "Default branch: %s\n", ctx.DefaultBranch)
	}
	if ctx.Branch != "" && ctx.DefaultBranch != "" {
		fmt.Fprintf(w, "Ahead/Behind:   %d / %d\n", ctx.Ahead, ctx.Behind)
	}
	fmt.Fprintf(w, "User:           %s (%s)\n", ctx.User.Slug, ctx.User.DisplayName)
	return nil
}

func contextFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Context] {
	p := format.New[backend.Context](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Context]{Name: "host", Header: "HOST", Extract: func(c backend.Context) any { return c.Host }})
	p.AddField(format.Field[backend.Context]{Name: "project", Header: "PROJECT", Extract: func(c backend.Context) any { return c.Project }})
	p.AddField(format.Field[backend.Context]{Name: "slug", Header: "SLUG", Extract: func(c backend.Context) any { return c.Slug }})
	p.AddField(format.Field[backend.Context]{Name: "branch", Header: "BRANCH", Extract: func(c backend.Context) any { return c.Branch }})
	p.AddField(format.Field[backend.Context]{Name: "default_branch", Header: "DEFAULT", Extract: func(c backend.Context) any { return c.DefaultBranch }})
	p.AddField(format.Field[backend.Context]{Name: "ahead", Header: "AHEAD", Extract: func(c backend.Context) any { return c.Ahead }})
	p.AddField(format.Field[backend.Context]{Name: "behind", Header: "BEHIND", Extract: func(c backend.Context) any { return c.Behind }})
	p.AddField(format.Field[backend.Context]{Name: "user", Header: "USER", Extract: func(c backend.Context) any {
		// gojq cannot traverse struct values; emit a map so `--jq .user.slug`
		// works the same as it would on the json.Marshal output.
		return map[string]any{
			"slug":         c.User.Slug,
			"display_name": c.User.DisplayName,
		}
	}})
	p.AddField(format.Field[backend.Context]{Name: "backend", Header: "BACKEND", Extract: func(c backend.Context) any { return c.Backend }})
	return p
}
