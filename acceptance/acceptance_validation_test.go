package acceptance

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestAcceptance_EnabledInvalidConfigurationFailsBeforeScenarios(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{
			name: "missing configuration",
			env: map[string]string{
				"BITBOTTLE_E2E":  "1",
				"BB_E2E_BACKEND": "server",
			},
			want: "BB_TOKEN",
		},
		{
			name: "unsafe target",
			env: map[string]string{
				"BITBOTTLE_E2E":  "1",
				"BB_E2E_BACKEND": "server",
				"BB_TOKEN":       "test-token-must-not-appear",
				"BB_HOST":        "bitbucket.example.test",
				"BB_E2E_REPO":    "PROJECT/production-sandbox",
				"BITBOTTLE_BIN":  binary,
				"BB_E2E_BRANCH":  "acceptance/test-branch",
				"BB_REVIEWER":    "reviewer",
			},
			want: "production safety guard",
		},
		{
			name: "cloud emits named scenario subtest",
			env:  map[string]string{"BITBOTTLE_E2E": "1", "BB_E2E_BACKEND": "cloud"},
			want: "TestAcceptance/cloud-pr-cli-mcp-lifecycle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestAcceptance$")
			cmd.Env = acceptanceTestEnv(tt.env)
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("enabled acceptance run unexpectedly succeeded")
			}
			if !strings.Contains(string(output), tt.want) {
				t.Fatalf("failure output did not contain %q", tt.want)
			}
			if strings.Contains(string(output), "test-token-must-not-appear") {
				t.Fatal("failure output exposed the token")
			}
		})
	}
}

func TestValidateAcceptanceConfig_RejectsEmptyScenarioSet(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{
		"BB_E2E_BACKEND": "cloud",
		"BB_TOKEN":       "test-token",
		"BB_HOST":        "bitbucket.org",
		"BB_E2E_REPO":    "workspace/sandbox",
		"BITBOTTLE_BIN":  binary,
		"BB_E2E_BRANCH":  "acceptance/test-branch",
		"BB_REVIEWER":    "reviewer",
	}
	tempDir := t.TempDir()

	_, err = validateAcceptanceConfig(func(key string) string { return values[key] }, tempDir)
	if err == nil || !strings.Contains(err.Error(), "no .txt or .txtar scenarios") {
		t.Fatalf("validateAcceptanceConfig() error = %v, want empty-scenario error", err)
	}
}

func TestValidateAcceptanceConfig_RequiresBackendHostMatch(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	base := map[string]string{
		"BB_TOKEN":      "test-token",
		"BB_E2E_REPO":   "workspace/sandbox",
		"BITBOTTLE_BIN": binary,
		"BB_E2E_BRANCH": "acceptance/test-branch",
		"BB_REVIEWER":   "reviewer",
	}
	for _, tt := range []struct {
		backend string
		host    string
	}{
		{backend: "cloud", host: "bitbucket.example.test"},
		{backend: "server", host: "bitbucket.org"},
	} {
		t.Run(tt.backend+"/"+tt.host, func(t *testing.T) {
			values := make(map[string]string, len(base)+2)
			for key, value := range base {
				values[key] = value
			}
			values["BB_E2E_BACKEND"] = tt.backend
			values["BB_HOST"] = tt.host
			_, err := validateAcceptanceConfig(func(key string) string { return values[key] }, "testdata/pr")
			if err == nil || !strings.Contains(err.Error(), "does not match") {
				t.Fatalf("validateAcceptanceConfig() error = %v, want backend-host mismatch", err)
			}
		})
	}
}

func acceptanceTestEnv(overrides map[string]string) []string {
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "BB_") || strings.HasPrefix(key, "BITBOTTLE_") {
			continue
		}
		env = append(env, entry)
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	return env
}
