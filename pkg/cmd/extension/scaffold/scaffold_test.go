package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/extension/scaffold"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
)

func TestScaffold_Go_CreatesExpectedFiles(t *testing.T) {
	dir := t.TempDir()
	f, out, _ := factorytest.New(t, factorytest.Opts{})

	cmd := scaffold.NewCmdScaffold(f)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"myplugin", "--lang", "go", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("scaffold go: %v", err)
	}

	root := filepath.Join(dir, "bitbottle-myplugin")
	for _, rel := range []string{
		"main.go",
		"go.mod",
		".github/workflows/release.yml",
		"README.md",
		"LICENSE",
	} {
		path := filepath.Join(root, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s: %v", rel, err)
		}
	}

	if !strings.Contains(out.String(), "bitbottle-myplugin") {
		t.Errorf("output %q missing project name", out.String())
	}
}

func TestScaffold_Bash_CreatesExpectedFiles(t *testing.T) {
	dir := t.TempDir()
	f, out, _ := factorytest.New(t, factorytest.Opts{})

	cmd := scaffold.NewCmdScaffold(f)
	cmd.SetOut(out)
	cmd.SetArgs([]string{"myplugin", "--lang", "bash", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("scaffold bash: %v", err)
	}

	root := filepath.Join(dir, "bitbottle-myplugin")
	for _, rel := range []string{
		"bitbottle-myplugin",
		".github/workflows/release.yml",
		"README.md",
		"LICENSE",
	} {
		path := filepath.Join(root, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s: %v", rel, err)
		}
	}

	// Bash entry point must be executable.
	info, err := os.Stat(filepath.Join(root, "bitbottle-myplugin"))
	if err != nil {
		t.Fatalf("stat bash script: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Errorf("bitbottle-myplugin is not executable: %v", info.Mode())
	}
}

func TestScaffold_DefaultLang_IsGo(t *testing.T) {
	dir := t.TempDir()
	f, _, _ := factorytest.New(t, factorytest.Opts{})

	cmd := scaffold.NewCmdScaffold(f)
	cmd.SetArgs([]string{"defaultlang", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("scaffold default lang: %v", err)
	}

	root := filepath.Join(dir, "bitbottle-defaultlang")
	if _, err := os.Stat(filepath.Join(root, "main.go")); err != nil {
		t.Errorf("expected main.go for default lang: %v", err)
	}
}

func TestScaffold_InvalidLang_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	f, _, _ := factorytest.New(t, factorytest.Opts{})

	cmd := scaffold.NewCmdScaffold(f)
	cmd.SetArgs([]string{"myplugin", "--lang", "python", "--dir", dir})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported language; got nil")
	}
	if !strings.Contains(err.Error(), "python") {
		t.Errorf("error %q missing unsupported lang name", err.Error())
	}
}

func TestScaffold_ExistingDir_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	f, _, _ := factorytest.New(t, factorytest.Opts{})

	// Pre-create the destination directory.
	if err := os.MkdirAll(filepath.Join(dir, "bitbottle-clash"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := scaffold.NewCmdScaffold(f)
	cmd.SetArgs([]string{"clash", "--dir", dir})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when destination already exists; got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %q missing 'already exists'", err.Error())
	}
}

func TestScaffold_TemplateVariablesSubstituted(t *testing.T) {
	dir := t.TempDir()
	f, _, _ := factorytest.New(t, factorytest.Opts{})

	cmd := scaffold.NewCmdScaffold(f)
	cmd.SetArgs([]string{"demo", "--lang", "go", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	root := filepath.Join(dir, "bitbottle-demo")

	mainGo, err := os.ReadFile(filepath.Join(root, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if !strings.Contains(string(mainGo), "bitbottle-demo") {
		t.Errorf("main.go missing BinaryName: %s", mainGo)
	}

	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}
	if !strings.Contains(string(readme), "bitbottle extension install OWNER/bitbottle-demo") {
		t.Errorf("README.md missing install instruction: %s", readme)
	}
}

func TestScaffold_BashScriptContent(t *testing.T) {
	dir := t.TempDir()
	f, _, _ := factorytest.New(t, factorytest.Opts{})

	cmd := scaffold.NewCmdScaffold(f)
	cmd.SetArgs([]string{"greet", "--lang", "bash", "--dir", dir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("scaffold bash: %v", err)
	}

	script, err := os.ReadFile(filepath.Join(dir, "bitbottle-greet", "bitbottle-greet"))
	if err != nil {
		t.Fatalf("read bash script: %v", err)
	}
	content := string(script)
	if !strings.HasPrefix(content, "#!/usr/bin/env bash") {
		t.Errorf("bash script missing shebang: %s", content)
	}
	if !strings.Contains(content, "bitbottle-greet") {
		t.Errorf("bash script missing BinaryName: %s", content)
	}
}
