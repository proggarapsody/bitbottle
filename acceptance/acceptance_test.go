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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/proggarapsody/bitbottle/internal/acceptancecloud"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
)

// TestAcceptance runs all acceptance testscript scripts under testdata/pr/.
// Ordinary test runs skip; an explicit live opt-in must have a complete,
// safe configuration and at least one discovered scenario.
func TestAcceptance(t *testing.T) {
	if os.Getenv("BITBOTTLE_E2E") != "1" {
		t.Skip("acceptance tests skipped: set BITBOTTLE_E2E=1 to run")
	}

	if strings.EqualFold(strings.TrimSpace(os.Getenv("BB_E2E_BACKEND")), "cloud") {
		t.Run(acceptancecloud.ScenarioID, func(t *testing.T) {
			report, err := acceptancecloud.Run(t.Context(), acceptancecloud.Config{
				BootstrapToken: os.Getenv("BB_CLOUD_BOOTSTRAP_TOKEN"), ActorToken: os.Getenv("BB_TOKEN"),
				Email: os.Getenv("BB_CLOUD_EMAIL"), Workspace: os.Getenv("BB_CLOUD_WORKSPACE"), Project: os.Getenv("BB_CLOUD_PROJECT"),
				ReportDir: os.Getenv("BB_ACCEPTANCE_REPORT_DIR"), Binary: os.Getenv("BITBOTTLE_BIN"),
			})
			if err != nil {
				t.Fatalf("Cloud acceptance %s: %v", report.Status, err)
			}
			if report.Status != "pass" {
				t.Fatalf("Cloud acceptance status = %s", report.Status)
			}
		})
		return
	}
	scripts, err := validateAcceptanceConfig(os.Getenv, filepath.Join("testdata", "pr"))
	if err != nil {
		t.Fatal(err)
	}

	testscript.Run(t, testscript.Params{
		Files: scripts,
		Setup: setup,
		Cmds:  customCmds(),
	})
}

func validateAcceptanceConfig(getenv func(string) string, scenarioDir string) ([]string, error) {
	backend := strings.ToLower(strings.TrimSpace(getenv("BB_E2E_BACKEND")))
	if backend != "cloud" && backend != "server" {
		return nil, fmt.Errorf("acceptance configuration: BB_E2E_BACKEND must be cloud or server")
	}

	required := []string{
		"BB_TOKEN",
		"BB_HOST",
		"BB_E2E_REPO",
		"BITBOTTLE_BIN",
		"BB_E2E_BRANCH",
		"BB_REVIEWER",
	}
	var missing []string
	for _, name := range required {
		if strings.TrimSpace(getenv(name)) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("acceptance configuration: required values are empty: %s", strings.Join(missing, ", "))
	}

	host := getenv("BB_HOST")
	if bbinstance.IsCloud(host, "") != (backend == "cloud") {
		return nil, fmt.Errorf("acceptance configuration: BB_E2E_BACKEND does not match the adapter selected by BB_HOST")
	}

	repo := getenv("BB_E2E_REPO")
	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 2 || repoParts[0] == "" || repoParts[1] == "" {
		return nil, fmt.Errorf("acceptance configuration: BB_E2E_REPO must be WORKSPACE/REPO or PROJECT/REPO")
	}
	if strings.Contains(strings.ToLower(repo), "prod") {
		return nil, fmt.Errorf("acceptance configuration: BB_E2E_REPO is blocked by the production safety guard")
	}

	if _, err := exec.LookPath(getenv("BITBOTTLE_BIN")); err != nil {
		return nil, fmt.Errorf("acceptance configuration: BITBOTTLE_BIN must point to an executable")
	}

	scripts, err := acceptanceScenarioFiles(scenarioDir)
	if err != nil {
		return nil, fmt.Errorf("acceptance configuration: cannot inspect scenario set: %w", err)
	}
	if len(scripts) == 0 {
		return nil, fmt.Errorf("acceptance configuration: no .txt or .txtar scenarios found")
	}
	return scripts, nil
}

func acceptanceScenarioFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var scripts []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || (!strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".txtar")) {
			continue
		}
		scripts = append(scripts, filepath.Join(dir, name))
	}
	return scripts, nil
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
