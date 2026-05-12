package format_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/proggarapsody/bitbottle/internal/format"
)

func TestWriteYAML_Roundtrip(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	in := map[string]any{
		"id":    42,
		"title": "Fix auth",
		"open":  true,
	}
	require.NoError(t, format.WriteYAML(&buf, in))

	var got map[string]any
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, 42, got["id"])
	assert.Equal(t, "Fix auth", got["title"])
	assert.Equal(t, true, got["open"])
}

func TestWriteYAML_Slice(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	in := []map[string]any{
		{"id": 1, "name": "a"},
		{"id": 2, "name": "b"},
	}
	require.NoError(t, format.WriteYAML(&buf, in))
	out := buf.String()
	assert.True(t, strings.Contains(out, "- "), "expected list dash, got: %s", out)
	assert.Contains(t, out, "name: a")
	assert.Contains(t, out, "name: b")
}
