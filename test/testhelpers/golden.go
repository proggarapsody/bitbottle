package testhelpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertGolden compares got against the contents of testdata/<name>.golden
// relative to the current test's package. When the BITBOTTLE_UPDATE_GOLDEN
// environment variable is set to a truthy value, the golden file is written
// instead of compared.
func AssertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if update := os.Getenv("BITBOTTLE_UPDATE_GOLDEN"); update == "1" || update == "true" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("testhelpers: mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("testhelpers: writing golden %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("testhelpers: reading golden %s: %v (set BITBOTTLE_UPDATE_GOLDEN=1 to create)", path, err)
	}
	assert.Equal(t, string(want), got, "golden mismatch for %s", name)
}
