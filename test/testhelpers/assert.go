package testhelpers

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertJSONArray asserts that got is a valid JSON array of length n.
func AssertJSONArray(t *testing.T, got string, n int) {
	t.Helper()
	var arr []any
	require.NoError(t, json.Unmarshal([]byte(got), &arr), "expected valid JSON array; got: %s", got)
	assert.Len(t, arr, n, "unexpected JSON array length")
}

// AssertNoOutput asserts that a buffer-like value produced no output.
func AssertNoOutput(t *testing.T, buf interface{ String() string }) {
	t.Helper()
	assert.Empty(t, buf.String(), "expected no output")
}

// AssertContainsAuthHint asserts the output contains an auth login hint for
// the given hostname.
func AssertContainsAuthHint(t *testing.T, got, hostname string) {
	t.Helper()
	if !strings.Contains(got, "auth login") {
		t.Errorf("expected output to contain auth login hint; got: %s", got)
		return
	}
	if hostname != "" && !strings.Contains(got, hostname) {
		t.Errorf("expected output to reference hostname %q; got: %s", hostname, got)
	}
}
