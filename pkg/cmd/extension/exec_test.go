package extension_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/extension"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
)

// writeExtension creates a minimal installed extension (manifest + stub binary)
// under configDir/extensions/<name>/.
func writeExtension(t *testing.T, configDir, name string) {
	t.Helper()
	binDir := filepath.Join(configDir, "extensions", name, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(binDir, name)
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	mf := map[string]any{"name": name, "local": true}
	data, _ := json.Marshal(mf)
	mfPath := filepath.Join(configDir, "extensions", name, "manifest.json")
	if err := os.WriteFile(mfPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestExecCmd_NotInstalled(t *testing.T) {
	configDir := t.TempDir()
	f, out, errOut := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"exec", "nonexistent"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error executing non-installed extension; got nil")
	}
}

func TestExecCmd_TokenInjectedInEnv(t *testing.T) {
	configDir := t.TempDir()
	writeExtension(t, configDir, "hello")

	// Record env that the runner sees by writing it to a temp file
	// (the stub runner can't modify variables in the test process).
	envFile := filepath.Join(t.TempDir(), "env.txt")

	f, out, errOut := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	// Set BB_TOKEN in the process env — exec.go reads it via os.Getenv.
	t.Setenv("BB_TOKEN", "test-token-abc")

	// Use a real binary that dumps its env to envFile.
	// We replace the extension binary with a shell script.
	binPath := filepath.Join(configDir, "extensions", "hello", "bin", "hello")
	script := "#!/bin/sh\nenv > " + envFile + "\n"
	if err := os.WriteFile(binPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"exec", "hello"})

	// Execute; ignore errors from the script itself (some CI envs might not have sh).
	_ = cmd.Execute()

	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Skip("shell script runner not available in this environment")
	}

	output := string(data)
	if !strings.Contains(output, "BB_TOKEN=test-token-abc") {
		t.Errorf("BB_TOKEN not injected; env output:\n%s", output)
	}
	if !strings.Contains(output, "BITBOTTLE_VERSION=") {
		t.Errorf("BITBOTTLE_VERSION not injected; env output:\n%s", output)
	}
}
