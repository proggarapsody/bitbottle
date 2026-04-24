// Package git wraps internal/run.Runner to provide git-specific operations.
package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aleksey/bitbottle/internal/run"
)

// Git provides git operations via the Runner.
type Git struct {
	runner run.Runner
}

// New constructs a Git wrapper around the given Runner.
func New(runner run.Runner) *Git {
	return &Git{runner: runner}
}

// CurrentBranch returns the name of the current branch.
func (g *Git) CurrentBranch() (string, error) {
	stdout, _, err := g.runner.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(stdout)
	if branch == "HEAD" {
		return "", fmt.Errorf("not on any branch (detached HEAD)")
	}
	return branch, nil
}

// RemoteURL returns the URL for the named remote.
func (g *Git) RemoteURL(name string) (string, error) {
	stdout, _, err := g.runner.Run("remote", "get-url", name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// HasUncommittedChanges returns true if the working tree is dirty.
func (g *Git) HasUncommittedChanges() (bool, error) {
	stdout, _, err := g.runner.Run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(stdout) != "", nil
}

// Fetch fetches a branch from remote.
func (g *Git) Fetch(remote, branch string) error {
	_, _, err := g.runner.Run("fetch", remote, branch)
	return err
}

// Checkout checks out a branch.
func (g *Git) Checkout(branch string) error {
	_, _, err := g.runner.Run("checkout", branch)
	return err
}

// Clone clones a repo URL into dir (empty string = default dir).
func (g *Git) Clone(url, dir string) error {
	args := []string{"clone", url}
	if dir != "" {
		args = append(args, dir)
	}
	_, _, err := g.runner.Run(args...)
	return err
}

// CommitsAhead returns number of commits in current branch ahead of base.
func (g *Git) CommitsAhead(base string) (int, error) {
	stdout, _, err := g.runner.Run("rev-list", "HEAD..."+base, "--count")
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(stdout))
	if err != nil {
		return 0, fmt.Errorf("parsing commit count: %w", err)
	}
	return n, nil
}
