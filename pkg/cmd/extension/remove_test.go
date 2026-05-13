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

func installFakeExtension(t *testing.T, extRoot, name string) {
	t.Helper()
	extDir := filepath.Join(extRoot, name)
	if err := os.MkdirAll(filepath.Join(extDir, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	mf := map[string]any{
		"name":    name,
		"repo":    "owner/bitbottle-" + name,
		"version": "v1.0.0",
		"local":   false,
	}
	data, _ := json.Marshal(mf)
	if err := os.WriteFile(filepath.Join(extDir, "manifest.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveCmd_Existing(t *testing.T) {
	configDir := t.TempDir()
	extRoot := filepath.Join(configDir, "extensions")
	if err := os.MkdirAll(extRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	installFakeExtension(t, extRoot, "hello")

	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"remove", "hello"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !strings.Contains(out.String(), "hello") {
		t.Errorf("output %q missing extension name", out.String())
	}
	if _, err := os.Stat(filepath.Join(extRoot, "hello")); !os.IsNotExist(err) {
		t.Errorf("extension dir still exists after remove")
	}
}

func TestRemoveCmd_NonExistent(t *testing.T) {
	configDir := t.TempDir()
	f, out, errOut := factorytest.New(t, factorytest.Opts{ConfigDir: configDir})
	f.ConfigDir = func() string { return configDir }

	cmd := extension.NewCmdExtension(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"remove", "nonexistent"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error removing non-existent extension; got nil")
	}
}
