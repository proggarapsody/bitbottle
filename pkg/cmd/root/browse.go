package root

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// hexHashRE matches a 7-40 character hexadecimal string (a commit hash).
var hexHashRE = regexp.MustCompile(`^[0-9a-fA-F]{7,40}$`)

// NewCmdBrowse returns the `bb browse [PROJECT/REPO] [TARGET]` command.
// Without TARGET: opens the repo web URL.
// Numeric TARGET: opens the PR web URL.
// 7-40 hex chars: opens the commit page.
// Anything else: opens src/<currentBranch>/<target> in the repo.
func NewCmdBrowse(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "browse [PROJECT/REPO] [TARGET]",
		Short: "Open a repository page in the browser",
		Long: `Open a Bitbucket page in the web browser.

PROJECT/REPO is optional; if omitted the current directory's repo is used.

TARGET can be:
  (none)         — open the repository home page
  PR_NUMBER      — open a specific pull request
  COMMIT_HASH    — open a specific commit (7-40 hex chars)
  PATH           — open src/<branch>/PATH in the repository`,
		Args: cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Split positional args: first is always the repo ref (if
			// present), second is the optional target within that repo.
			var repoArgs []string
			var target string

			switch len(args) {
			case 0:
				// nothing — both repo and target come from context
			case 1:
				// Could be "PROJECT/REPO" (repo only) or a target when
				// the repo comes from the CWD. We treat arg[0] as a repo
				// ref when it contains exactly one "/" separator and
				// neither part looks like a PR number, hex hash, or path
				// with an extension — i.e. when it matches PROJECT/REPO.
				repoArgs = []string{args[0]}
			case 2:
				repoArgs = []string{args[0]}
				target = args[1]
			}

			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			// Get the repo to obtain its web URL
			repo, err := client.GetRepo(ref.Project, ref.Slug)
			if err != nil {
				return err
			}
			repoWebURL := repo.WebURL

			// No target: open repo home
			if target == "" {
				if repoWebURL == "" {
					return fmt.Errorf("no web URL available for this repository")
				}
				return f.Browser.Browse(repoWebURL)
			}

			// Numeric: open PR
			if prID, err := strconv.Atoi(target); err == nil {
				p, err := client.GetPR(ref.Project, ref.Slug, prID)
				if err != nil {
					return err
				}
				if p.WebURL == "" {
					return fmt.Errorf("no web URL available for PR #%d", prID)
				}
				return f.Browser.Browse(p.WebURL)
			}

			// 7-40 hex chars: open commit page.
			// Server repo WebURL ends in "/browse"; commit URL strips that suffix.
			// Cloud repo WebURL ends in "/<slug>"; commit URL appends "/commits/".
			if hexHashRE.MatchString(target) {
				base := strings.TrimRight(repoWebURL, "/")
				var commitURL string
				if strings.HasSuffix(base, "/browse") {
					commitURL = strings.TrimSuffix(base, "/browse") + "/commits/" + target
				} else {
					commitURL = base + "/commits/" + target
				}
				return f.Browser.Browse(commitURL)
			}

			// Anything else: open the file/path in the source browser.
			// Server: {repoWebURL}/{path}?at=refs/heads/{branch}
			// Cloud:  {repoWebURL}/src/{branch}/{path}
			branch := currentBranchForBrowse(f)
			if branch == "" {
				branch = "main"
			}
			base := strings.TrimRight(repoWebURL, "/")
			var srcURL string
			if strings.HasSuffix(base, "/browse") {
				srcURL = base + "/" + strings.TrimLeft(target, "/") + "?at=refs/heads/" + branch
			} else {
				srcURL = base + "/src/" + branch + "/" + strings.TrimLeft(target, "/")
			}
			return f.Browser.Browse(srcURL)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func currentBranchForBrowse(f *factory.Factory) string {
	runner := f.GitRunner()
	g := git.New(runner)
	branch, err := g.CurrentBranch()
	if err != nil {
		return ""
	}
	return branch
}
