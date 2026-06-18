// Package acceptance runs real-backend acceptance tests against a live
// Bitbucket Server or Cloud sandbox.
//
// Tests in this package are gated by the BITBOTTLE_E2E=1 environment
// variable so they NEVER run in ordinary CI and never contact the real API
// unless explicitly opted in.
//
// To run locally:
//
//	BITBOTTLE_E2E=1 BB_E2E_REPO=PROJECT/my-throwaway-repo \
//	BB_TOKEN=<token> BB_HOST=<host> \
//	BITBOTTLE_BIN=/path/to/bitbottle \
//	go test ./acceptance/... -v -timeout 20m -run TestAcceptance
package acceptance

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestAcceptance runs all acceptance testscript scripts under testdata/pr/.
// It skips cleanly (t.Skip, not t.Fatal) when:
//   - BITBOTTLE_E2E is not set to "1"
//   - BB_E2E_REPO is absent
//   - BB_E2E_REPO contains "prod" (safety guard against accidentally hitting production)
func TestAcceptance(t *testing.T) {
	if os.Getenv("BITBOTTLE_E2E") != "1" {
		t.Skip("acceptance tests skipped: set BITBOTTLE_E2E=1 to run")
	}

	repo := os.Getenv("BB_E2E_REPO")
	if repo == "" {
		t.Skip("acceptance tests skipped: BB_E2E_REPO is not set")
	}

	if strings.Contains(strings.ToLower(repo), "prod") {
		t.Skipf("acceptance tests skipped: BB_E2E_REPO=%q contains 'prod' — refusing to run against a production repo", repo)
	}

	testscript.Run(t, testscript.Params{
		Dir:   "testdata/pr",
		Setup: setup,
		Cmds:  customCmds(),
	})
}

// setup passes through all BB_* and BITBOTTLE_* env vars from the real
// process environment into the testscript engine. Unlike the hermetic
// test/script/script_test.go, we deliberately do NOT scrub these vars —
// the whole point is to use real credentials against a real backend.
func setup(env *testscript.Env) error {
	// Pass through all BB_* and BITBOTTLE_* env vars from the host process.
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(k, "BB_") || strings.HasPrefix(k, "BITBOTTLE_") {
			env.Vars = append(env.Vars, kv)
		}
	}

	return nil
}

// prNumberRe matches the first "#NNN" sequence in a line and captures the digits.
var prNumberRe = regexp.MustCompile(`#(\d+)`)

// customCmds returns the extra testscript commands available in acceptance scripts.
func customCmds() map[string]func(ts *testscript.TestScript, neg bool, args []string) {
	return map[string]func(ts *testscript.TestScript, neg bool, args []string){
		// stdout2env KEY — captures the last non-empty line of stdout and exports
		// it as a testscript env var with the given KEY name.
		//
		// If the line contains a "#N" pattern (e.g. "Created pull request #42:
		// Acceptance: reviewer-safety seed PR"), extracts only the digits after
		// the first "#" — so PR_NUMBER will be "42", not the full line.
		//
		// Usage in a .txtar script:
		//
		//   exec $BITBOTTLE_BIN pr create ...
		//   stdout2env PR_NUMBER
		//   exec $BITBOTTLE_BIN pr view $PR_NUMBER
		"stdout2env": func(ts *testscript.TestScript, neg bool, args []string) {
			if len(args) != 1 {
				ts.Fatalf("stdout2env: usage: stdout2env KEY")
			}
			key := args[0]
			out := ts.ReadFile("stdout")
			// Trim trailing whitespace and take the last non-empty line.
			lines := strings.Split(strings.TrimRight(out, "\r\n"), "\n")
			var last string
			for i := len(lines) - 1; i >= 0; i-- {
				line := strings.TrimSpace(lines[i])
				if line != "" {
					last = line
					break
				}
			}
			if last == "" {
				ts.Fatalf("stdout2env: stdout is empty, cannot set %s", key)
			}
			// If the line contains "#N", extract just the digits (e.g. "42" from
			// "Created pull request #42: ...").
			if m := prNumberRe.FindStringSubmatch(last); m != nil {
				last = m[1]
			}
			ts.Setenv(key, last)
		},
	}
}
