package format

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests live in `package format` (not _test) so we can exercise the
// unexported template-func helpers directly. The exported tests for
// WriteTemplate could live in either package; keeping them here is fine.

func TestWriteTemplate_RendersFields(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := map[string]any{"id": 42, "title": "Fix auth"}
	require.NoError(t, WriteTemplate(&buf, `{{.id}}: {{.title}}`, data))
	assert.Equal(t, "42: Fix auth", buf.String())
}

func TestWriteTemplate_ParseError(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := WriteTemplate(&buf, `{{.id`, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--template")
}

func TestTemplateFunc_Color(t *testing.T) {
	t.Parallel()
	got := colorFunc("red", "danger")
	assert.Equal(t, "\x1b[31mdanger\x1b[0m", got)

	// unknown color leaves text untouched
	assert.Equal(t, "x", colorFunc("not-a-color", "x"))
}

func TestTemplateFunc_Truncate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		max  int
		in   string
		want string
	}{
		{"shorter", 10, "hi", "hi"},
		{"equal", 2, "hi", "hi"},
		{"longer", 5, "abcdefgh", "abcd…"},
		{"one", 1, "abc", "…"},
		{"zero", 0, "abc", ""},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, truncateFunc(c.max, c.in))
		})
	}
}

func TestTemplateFunc_Timeago(t *testing.T) {
	t.Parallel()
	now := time.Now()
	cases := []struct {
		name string
		t    time.Time
		want string
	}{
		{"zero", time.Time{}, ""},
		{"just_now", now.Add(-10 * time.Second), "just now"},
		{"one_minute", now.Add(-1 * time.Minute), "1 minute ago"},
		{"five_minutes", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"two_hours", now.Add(-2 * time.Hour), "2 hours ago"},
		{"three_days", now.Add(-3 * 24 * time.Hour), "3 days ago"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, timeagoFunc(c.t))
		})
	}
}

func TestTemplateFunc_Pluck(t *testing.T) {
	t.Parallel()
	items := []any{
		map[string]any{"id": 1, "name": "a"},
		map[string]any{"id": 2, "name": "b"},
		map[string]any{"id": 3, "name": "c"},
	}
	assert.Equal(t, []any{"a", "b", "c"}, pluckFunc("name", items))
}

func TestTemplateFunc_Join(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "a, b, c", joinFunc(", ", []any{"a", "b", "c"}))
	assert.Equal(t, "1-2-3", joinFunc("-", []any{1, 2, 3}))
}

func TestTemplateFunc_Tablerender(t *testing.T) {
	t.Parallel()
	rows := [][]string{
		{"ID", "TITLE"},
		{"1", "Fix"},
		{"2", "Bump"},
	}
	got := tablerenderFunc(rows)
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	assert.Equal(t, []string{"ID\tTITLE", "1\tFix", "2\tBump"}, lines)
}

func TestTemplateFunc_Hyperlink(t *testing.T) {
	t.Parallel()
	got := hyperlinkFunc("https://example.com", "click")
	assert.Contains(t, got, "\x1b]8;;https://example.com\x1b\\")
	assert.Contains(t, got, "click")
}

func TestWriteTemplate_RangeAndColor(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	data := []map[string]any{
		{"name": "a"},
		{"name": "b"},
	}
	require.NoError(t, WriteTemplate(&buf, `{{range .}}{{color "green" .name}} {{end}}`, data))
	out := buf.String()
	assert.Contains(t, out, "\x1b[32ma\x1b[0m")
	assert.Contains(t, out, "\x1b[32mb\x1b[0m")
}
