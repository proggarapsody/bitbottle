// Package git wraps internal/run.Runner to provide git-specific operations.
package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/proggarapsody/bitbottle/internal/run"
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

// GetConfig reads a value from the local repository's git config. A missing
// key returns ("", nil) — git config exits non-zero in that case but callers
// should treat absence as "fall back to other inference" rather than as an
// error condition.
func (g *Git) GetConfig(key string) (string, error) {
	stdout, _, err := g.runner.Run("config", "--local", "--get", key)
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(stdout), nil
}

// SetConfig writes a key-value pair to the local repository's git config.
// Used by `repo set-default` to pin a chosen Bitbucket coordinate so future
// commands in this checkout do not consult the git remote.
func (g *Git) SetConfig(key, value string) error {
	_, _, err := g.runner.Run("config", "--local", key, value)
	return err
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
