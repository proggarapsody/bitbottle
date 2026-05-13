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

func writeManifest(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestListCmd_NoExtensions(t *testing.T) {
	configDir := t.TempDir()
	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out.String(), "No extensions installed") {
		t.Errorf("output %q missing empty message", out.String())
	}
}

func TestListCmd_OneRemoteOneLocal(t *testing.T) {
	configDir := t.TempDir()
	extRoot := filepath.Join(configDir, "extensions")

	// Remote extension.
	remoteDir := filepath.Join(extRoot, "hello")
	if err := os.MkdirAll(filepath.Join(remoteDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, filepath.Join(remoteDir, "manifest.json"), map[string]any{
		"name":    "hello",
		"repo":    "owner/bitbottle-hello",
		"version": "v1.2.3",
		"local":   false,
	})

	// Local extension.
	localDir := filepath.Join(extRoot, "greet")
	if err := os.MkdirAll(filepath.Join(localDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, filepath.Join(localDir, "manifest.json"), map[string]any{
		"name":       "greet",
		"local":      true,
		"local_path": "/src/bitbottle-greet",
	})

	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "hello") {
		t.Errorf("output missing 'hello': %q", got)
	}
	if !strings.Contains(got, "v1.2.3") {
		t.Errorf("output missing version 'v1.2.3': %q", got)
	}
	if !strings.Contains(got, "greet") {
		t.Errorf("output missing 'greet': %q", got)
	}
	if !strings.Contains(got, "(local)") {
		t.Errorf("output missing '(local)': %q", got)
	}
}
