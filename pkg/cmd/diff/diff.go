// Package diff implements the top-level `bb diff` command.
// It shows a unified diff or a summary (--stat) between two refs in a
// Bitbucket repository.
package diff

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdregistry"
)

func init() {
	cmdregistry.Register(NewCmdDiff)
}

// NewCmdDiff builds the top-level `diff` command.
func NewCmdDiff(f *factory.Factory) *cobra.Command {
	var (
		hostname string
		stat     bool
	)

	cmd := &cobra.Command{
		Use:   "diff REF1[..REF2] [PROJECT/REPO]",
		Short: "Show diff between two refs",
		Long: `Show the diff between two refs (branches, tags, or commit hashes).

Refs can be supplied as a single "REF1..REF2" argument or as two separate
positional arguments "REF1 REF2". An optional PROJECT/REPO argument overrides
the repository resolved from the current directory.

Use --stat to print a summary of changed files instead of the full diff.`,
		Args: cobra.RangeArgs(1, 3),
		Annotations: map[string]string{
			"help:arguments": `REF1[..REF2] can be branch names, tags, or commit hashes.
When using two-argument form, REF1 and REF2 are supplied separately.
PROJECT/REPO is optional and overrides the repository from the current directory.`,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			from, to, repoArg, err := parseArgs(args)
			if err != nil {
				return err
			}

			var repoArgs []string
			if repoArg != "" {
				repoArgs = []string{repoArg}
			}
			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			dc, err := backend.AsDiffClient(client, ref.Host)
			if err != nil {
				return err
			}

			out := f.IOStreams.Out

			if stat {
				diffStat, err := dc.GetDiffStat(ref.Project, ref.Slug, from, to)
				if err != nil {
					return err
				}
				printStat(out, diffStat)
				return nil
			}

			text, err := dc.GetDiff(ref.Project, ref.Slug, from, to)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(out, text)
			return err
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().BoolVar(&stat, "stat", false, "Show summary of changed files instead of full diff")

	return cmd
}

// parseArgs interprets the positional arguments and returns (from, to, repoArg).
// Accepted forms:
//
//	diff A..B          → from=A, to=B, repoArg=""
//	diff A..B REPO     → from=A, to=B, repoArg=REPO
//	diff A B           → from=A, to=B, repoArg=""
//	diff A B REPO      → from=A, to=B, repoArg=REPO
func parseArgs(args []string) (from, to, repoArg string, err error) {
	switch len(args) {
	case 1:
		// Must be "A..B"
		parts := strings.SplitN(args[0], "..", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", "", fmt.Errorf("expected REF1..REF2 when a single argument is given, got %q", args[0])
		}
		return parts[0], parts[1], "", nil
	case 2:
		// Could be "A..B REPO" or "A B"
		if strings.Contains(args[0], "..") {
			parts := strings.SplitN(args[0], "..", 2)
			if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
				return parts[0], parts[1], args[1], nil
			}
		}
		return args[0], args[1], "", nil
	case 3:
		// "A B REPO"
		return args[0], args[1], args[2], nil
	default:
		return "", "", "", fmt.Errorf("unexpected number of arguments: %d", len(args))
	}
}

// printStat writes the diffstat summary to w in the format:
//
//	N files changed, X insertions(+), Y deletions(-)
//	  M  path (+X/-Y)
func printStat(out io.Writer, stat backend.DiffStat) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d file", stat.FilesChanged)
	if stat.FilesChanged != 1 {
		sb.WriteString("s")
	}
	sb.WriteString(" changed")
	if stat.Additions > 0 {
		fmt.Fprintf(&sb, ", %d insertion", stat.Additions)
		if stat.Additions != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("(+)")
	}
	if stat.Deletions > 0 {
		fmt.Fprintf(&sb, ", %d deletion", stat.Deletions)
		if stat.Deletions != 1 {
			sb.WriteString("s")
		}
		sb.WriteString("(-)")
	}
	sb.WriteString("\n")
	for _, entry := range stat.Files {
		var prefix string
		switch entry.Status {
		case "added":
			prefix = "A"
		case "deleted":
			prefix = "D"
		case "renamed":
			prefix = "R"
		default:
			prefix = "M"
		}
		fmt.Fprintf(&sb, "  %s  %s   (+%d/-%d)\n", prefix, entry.Path, entry.Additions, entry.Deletions)
	}
	_, _ = fmt.Fprint(out, sb.String())
}
