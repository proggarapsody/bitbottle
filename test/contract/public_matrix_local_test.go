package contract_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
)

type forbiddenLocalKeyring struct{ t *testing.T }

func (k forbiddenLocalKeyring) Get(string, string) (string, error) {
	k.t.Fatal("local command accessed the keyring")
	return "", nil
}
func (k forbiddenLocalKeyring) Set(string, string, string) error {
	k.t.Fatal("local command accessed the keyring")
	return nil
}
func (k forbiddenLocalKeyring) Delete(string, string) error {
	k.t.Fatal("local command accessed the keyring")
	return nil
}

// Each call creates a fresh public command and config loader. Any attempt to
// reach a provider or HTTP transport fails the test immediately.
func localCLI(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	f, out, _ := factorytest.New(t, factorytest.Opts{ConfigDir: dir})
	f.Keyring = forbiddenLocalKeyring{t}
	f.Backend = func(string) (backend.Client, error) {
		t.Fatal("local command requested a remote backend")
		return nil, nil
	}
	f.HTTPClient = func(string) (factory.HTTPClient, error) {
		t.Fatal("local command requested an HTTP client")
		return nil, nil
	}
	cmd := root.NewCmdRoot(f)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func localRecord(t *testing.T, operation, outcome string, assertions, inputs []string) {
	t.Helper()
	qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle " + operation,
		Backend: "local", Outcome: outcome, Assertions: assertions, Flags: inputs})
}

func TestPublicMatrixLocalAlias(t *testing.T) {
	for _, operation := range []string{"set", "list", "delete"} {
		t.Run(operation+"/success", func(t *testing.T) {
			dir := t.TempDir()
			_, err := localCLI(t, dir, "alias", "set", "qa", "pr list --author @me")
			require.NoError(t, err)
			var inputs []string
			switch operation {
			case "set":
				inputs = []string{"name", "expansion"}
				got, err := localCLI(t, dir, "alias", "list")
				require.NoError(t, err)
				require.Equal(t, "qa: pr list --author @me\n", got)
			case "list":
				got, err := localCLI(t, dir, "alias", "list")
				require.NoError(t, err)
				require.Equal(t, "qa: pr list --author @me\n", got)
			case "delete":
				inputs = []string{"name"}
				_, err := localCLI(t, dir, "alias", "delete", "qa")
				require.NoError(t, err)
				got, err := localCLI(t, dir, "alias", "list")
				require.NoError(t, err)
				require.Empty(t, got)
			}
			localRecord(t, "alias "+operation, "success", []string{"output", "effect", "input"}, inputs)
		})
		t.Run(operation+"/error", func(t *testing.T) {
			dir := t.TempDir()
			var args []string
			var inputs []string
			switch operation {
			case "set":
				args = []string{"alias", "set", "pr", "repo list"}
				inputs = []string{"name", "expansion"}
			case "delete":
				args = []string{"alias", "delete", "absent"}
				inputs = []string{"name"}
			case "list":
				require.NoError(t, os.WriteFile(filepath.Join(dir, "aliases.yml"), []byte("aliases: ["), 0o600))
				args = []string{"alias", "list"}
			}
			path := filepath.Join(dir, "aliases.yml")
			before, readErr := os.ReadFile(path)
			require.True(t, readErr == nil || os.IsNotExist(readErr))
			got, err := localCLI(t, dir, args...)
			require.Error(t, err)
			require.Empty(t, got)
			switch operation {
			case "set":
				require.ErrorContains(t, err, "shadow")
			case "delete":
				require.ErrorContains(t, err, "no alias named")
			case "list":
				require.ErrorContains(t, err, "yaml")
			}
			after, readErr := os.ReadFile(path)
			require.True(t, readErr == nil || os.IsNotExist(readErr))
			require.Equal(t, before, after)
			localRecord(t, "alias "+operation, "error", []string{"error", "effect", "input"}, inputs)
		})
	}
}

func TestPublicMatrixLocalConfig(t *testing.T) {
	for _, operation := range []string{"set", "get", "list"} {
		t.Run(operation+"/success", func(t *testing.T) {
			dir := t.TempDir()
			_, err := localCLI(t, dir, "config", "set", "git_protocol", "https", "--host", "bb.example.test")
			require.NoError(t, err)
			var inputs []string
			switch operation {
			case "set":
				inputs = []string{"key", "value", "--host"}
				got, err := localCLI(t, dir, "config", "get", "git_protocol", "--host", "bb.example.test")
				require.NoError(t, err)
				require.Equal(t, "https\n", got)
			case "get":
				inputs = []string{"key", "--host"}
				got, err := localCLI(t, dir, "config", "get", "git_protocol", "--host", "bb.example.test")
				require.NoError(t, err)
				require.Equal(t, "https\n", got)
			case "list":
				got, err := localCLI(t, dir, "config", "list")
				require.NoError(t, err)
				require.Equal(t, "bb.example.test.git_protocol=https\n", got)
			}
			localRecord(t, "config "+operation, "success", []string{"output", "effect", "input"}, inputs)
		})
		t.Run(operation+"/error", func(t *testing.T) {
			dir := t.TempDir()
			var args []string
			var inputs []string
			switch operation {
			case "set":
				args = []string{"config", "set", "editor", "vim", "--host", "bb.example.test"}
				inputs = []string{"key", "value", "--host"}
			case "get":
				args = []string{"config", "get", "missing"}
				inputs = []string{"key"}
			case "list":
				require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte("[broken"), 0o600))
				args = []string{"config", "list"}
			}
			path := filepath.Join(dir, "config.yml")
			before, readErr := os.ReadFile(path)
			require.True(t, readErr == nil || os.IsNotExist(readErr))
			got, err := localCLI(t, dir, args...)
			require.Error(t, err)
			require.Empty(t, got)
			switch operation {
			case "set":
				require.ErrorContains(t, err, "cannot be set per-host")
			case "get":
				require.ErrorContains(t, err, "no value set")
			case "list":
				require.ErrorContains(t, err, "yaml")
			}
			after, readErr := os.ReadFile(path)
			require.True(t, readErr == nil || os.IsNotExist(readErr))
			require.Equal(t, before, after)
			localRecord(t, "config "+operation, "error", []string{"error", "effect", "input"}, inputs)
		})
	}
}

func TestPublicMatrixLocalCompletion(t *testing.T) {
	markers := map[string]string{
		"bash":       "# bash completion for bitbottle",
		"zsh":        "#compdef bitbottle",
		"fish":       "function __bitbottle_debug",
		"powershell": "# powershell completion for bitbottle",
	}
	positional := make(map[string]string, len(markers))
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			got, err := localCLI(t, t.TempDir(), "completion", shell)
			require.NoError(t, err)
			require.Contains(t, got, markers[shell])
			positional[shell] = got
		})
	}
	got, err := localCLI(t, t.TempDir(), "completion", "--shell", "bash")
	require.NoError(t, err)
	require.Equal(t, positional["bash"], got)
	got, err = localCLI(t, t.TempDir(), "completion", "-s", "zsh")
	require.NoError(t, err)
	require.Equal(t, positional["zsh"], got)
	// Completion's effect is the emitted shell program; it is produced
	// locally and has no persistent state.
	localRecord(t, "completion", "success", []string{"output", "effect", "input"}, []string{"bash|zsh|fish|powershell", "--shell", "-s"})
	for _, args := range [][]string{{"completion"}, {"completion", "tcsh"}} {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yml")
		before := []byte("editor: vim\n")
		require.NoError(t, os.WriteFile(path, before, 0o600))
		got, err := localCLI(t, dir, args...)
		require.Error(t, err)
		require.Empty(t, got)
		if len(args) == 1 {
			require.ErrorContains(t, err, "shell is required")
		} else {
			require.ErrorContains(t, err, "tcsh")
		}
		after, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		require.Equal(t, before, after)
	}
	localRecord(t, "completion", "error", []string{"error", "effect", "input"}, []string{"bash|zsh|fish|powershell"})
}
