package git_test

import (
	"errors"
	"testing"

	"github.com/aleksey/bitbottle/git"
	"github.com/aleksey/bitbottle/test/testhelpers"
)

func TestGit_CurrentBranch(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "feat/login\n"},
	)
	g := git.New(runner)

	branch, err := g.CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch: unexpected error: %v", err)
	}
	if branch != "feat/login" {
		t.Errorf("CurrentBranch = %q, want %q", branch, "feat/login")
	}
	runner.AssertCalled(t, "rev-parse", "--abbrev-ref", "HEAD")
}

func TestGit_CurrentBranch_Detached(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "HEAD\n"},
	)
	g := git.New(runner)

	_, err := g.CurrentBranch()
	if err == nil {
		t.Fatal("CurrentBranch: expected error on detached HEAD, got nil")
	}
	runner.AssertCalled(t, "rev-parse", "--abbrev-ref", "HEAD")
}

func TestGit_RemoteURL(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "git@git.example.com:proj/repo.git\n"},
	)
	g := git.New(runner)

	url, err := g.RemoteURL("origin")
	if err != nil {
		t.Fatalf("RemoteURL: unexpected error: %v", err)
	}
	if url != "git@git.example.com:proj/repo.git" {
		t.Errorf("RemoteURL = %q, want %q", url, "git@git.example.com:proj/repo.git")
	}
	runner.AssertCalled(t, "remote", "get-url", "origin")
}

func TestGit_RemoteURL_NoRemote(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Err: errors.New("fatal: No such remote 'origin'")},
	)
	g := git.New(runner)

	_, err := g.RemoteURL("origin")
	if err == nil {
		t.Fatal("RemoteURL: expected error when remote does not exist, got nil")
	}
	runner.AssertCalled(t, "remote", "get-url", "origin")
}

func TestGit_HasUncommittedChanges_Clean(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: ""},
	)
	g := git.New(runner)

	dirty, err := g.HasUncommittedChanges()
	if err != nil {
		t.Fatalf("HasUncommittedChanges: unexpected error: %v", err)
	}
	if dirty {
		t.Errorf("HasUncommittedChanges = true, want false for clean tree")
	}
	runner.AssertCalled(t, "status", "--porcelain")
}

func TestGit_HasUncommittedChanges_Dirty(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: " M file.go\n?? new.go\n"},
	)
	g := git.New(runner)

	dirty, err := g.HasUncommittedChanges()
	if err != nil {
		t.Fatalf("HasUncommittedChanges: unexpected error: %v", err)
	}
	if !dirty {
		t.Errorf("HasUncommittedChanges = false, want true for dirty tree")
	}
	runner.AssertCalled(t, "status", "--porcelain")
}

func TestGit_Fetch(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{},
	)
	g := git.New(runner)

	if err := g.Fetch("origin", "feat/login"); err != nil {
		t.Fatalf("Fetch: unexpected error: %v", err)
	}
	runner.AssertCalled(t, "fetch", "origin", "feat/login")
}

func TestGit_Checkout(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{},
	)
	g := git.New(runner)

	if err := g.Checkout("feat/login"); err != nil {
		t.Fatalf("Checkout: unexpected error: %v", err)
	}
	runner.AssertCalled(t, "checkout", "feat/login")
}

func TestGit_Clone_NoDir(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{},
	)
	g := git.New(runner)

	url := "https://git.example.com/scm/proj/repo.git"
	if err := g.Clone(url, ""); err != nil {
		t.Fatalf("Clone: unexpected error: %v", err)
	}
	runner.AssertCalled(t, "clone", url)

	// Ensure no dir arg was appended.
	if len(runner.Calls) != 1 {
		t.Fatalf("expected exactly 1 call, got %d", len(runner.Calls))
	}
	args := runner.Calls[0].Args
	if len(args) != 2 || args[0] != "clone" || args[1] != url {
		t.Errorf("Clone args = %v, want [clone %s]", args, url)
	}
}

func TestGit_Clone_WithDir(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{},
	)
	g := git.New(runner)

	url := "https://git.example.com/scm/proj/repo.git"
	dir := "my-repo"
	if err := g.Clone(url, dir); err != nil {
		t.Fatalf("Clone: unexpected error: %v", err)
	}
	runner.AssertCalled(t, "clone", url, dir)
}

func TestGit_Clone_Failure(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("authentication failed")
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Err: wantErr},
	)
	g := git.New(runner)

	err := g.Clone("https://git.example.com/scm/proj/repo.git", "")
	if err == nil {
		t.Fatal("Clone: expected error, got nil")
	}
}

func TestGit_CommitsAhead_Zero(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "0\n"},
	)
	g := git.New(runner)

	n, err := g.CommitsAhead("main")
	if err != nil {
		t.Fatalf("CommitsAhead: unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("CommitsAhead = %d, want 0", n)
	}
	runner.AssertCalled(t, "rev-list", "HEAD...main", "--count")
}

func TestGit_CommitsAhead_Some(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "3\n"},
	)
	g := git.New(runner)

	n, err := g.CommitsAhead("main")
	if err != nil {
		t.Fatalf("CommitsAhead: unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("CommitsAhead = %d, want 3", n)
	}
	runner.AssertCalled(t, "rev-list", "HEAD...main", "--count")
}
