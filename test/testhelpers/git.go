package testhelpers

import (
	"strings"
	"sync"
	"testing"
)

// Call records a single invocation of the FakeRunner.
type Call struct {
	Args        []string
	Interactive bool
}

// RunResponse is a canned result for a single Run/RunInteractive invocation.
type RunResponse struct {
	Stdout, Stderr string
	Err            error
}

// FakeRunner is a test double for a git command runner.
type FakeRunner struct {
	mu        sync.Mutex
	responses []RunResponse
	Calls     []Call
	// BitbottleConfig overrides answers for `git config --local --get
	// bitbottle.<key>` lookups. Tests that exercise the pinned-default-repo
	// path populate this map; legacy tests leave it nil and bitbottle.* reads
	// resolve to "missing" without consuming a queued response.
	BitbottleConfig map[string]string
}

// NewFakeRunner constructs a FakeRunner seeded with the provided responses.
// Responses are consumed in order; if the runner is called more times than
// there are responses the extras return zero values.
func NewFakeRunner(responses ...RunResponse) *FakeRunner {
	return &FakeRunner{responses: responses}
}

// Run records the call and returns the next canned response.
//
// Reads of bitbottle.* keys via `git config --local --get bitbottle.<name>`
// are answered as "missing" (empty stdout, no error) without consuming a
// queued response. This lets pre-existing tests that seed responses only for
// the legacy git-remote inference path continue to work after the factory
// gained a `bitbottle repo set-default` consult step. Tests that intentionally
// exercise the pinned-default path supply explicit responses; their first
// `bitbottle.host` lookup matches the queued stdout before this shortcut
// fires.
func (r *FakeRunner) Run(args ...string) (string, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	argsCopy := append([]string(nil), args...)
	r.Calls = append(r.Calls, Call{Args: argsCopy})
	if isBitbottleConfigGet(args) {
		if v, ok := r.BitbottleConfig[args[3]]; ok {
			return v, "", nil
		}
		return "", "", nil
	}
	return r.next()
}

// isBitbottleConfigGet reports whether args is a `git config --local --get
// bitbottle.<key>` invocation.
func isBitbottleConfigGet(args []string) bool {
	if len(args) != 4 {
		return false
	}
	if args[0] != "config" || args[1] != "--local" || args[2] != "--get" {
		return false
	}
	return strings.HasPrefix(args[3], "bitbottle.")
}

// RunInteractive records the call (flagged as interactive) and returns only
// the error from the next canned response.
func (r *FakeRunner) RunInteractive(args ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	argsCopy := append([]string(nil), args...)
	r.Calls = append(r.Calls, Call{Args: argsCopy, Interactive: true})
	_, _, err := r.next()
	return err
}

// next must be called with r.mu held.
func (r *FakeRunner) next() (string, string, error) {
	if len(r.responses) == 0 {
		return "", "", nil
	}
	resp := r.responses[0]
	r.responses = r.responses[1:]
	return resp.Stdout, resp.Stderr, resp.Err
}

// AssertCalled fails t if no recorded call matches the supplied args exactly.
func (r *FakeRunner) AssertCalled(t testing.TB, args ...string) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.Calls {
		if argsEqual(c.Args, args) {
			return
		}
	}
	t.Errorf("expected FakeRunner to be called with %v; calls were %v", args, r.renderCalls())
}

// AssertNotCalled fails t if any recorded call matches the supplied args.
func (r *FakeRunner) AssertNotCalled(t testing.TB, args ...string) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.Calls {
		if argsEqual(c.Args, args) {
			t.Errorf("expected FakeRunner NOT to be called with %v; calls were %v", args, r.renderCalls())
			return
		}
	}
}

func (r *FakeRunner) renderCalls() []string {
	out := make([]string, 0, len(r.Calls))
	for _, c := range r.Calls {
		out = append(out, strings.Join(c.Args, " "))
	}
	return out
}

func argsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
