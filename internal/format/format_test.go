package format_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/format"
)

type testItem struct {
	ID    int
	Title string
	State string
}

func newPrinter(w *bytes.Buffer, isTTY bool, cfg format.OutputConfig) *format.Printer[testItem] {
	p := format.New[testItem](w, isTTY, cfg)
	p.AddField(format.Field[testItem]{Name: "id", Header: "ID", Extract: func(i testItem) any { return i.ID }})
	p.AddField(format.Field[testItem]{Name: "title", Header: "TITLE", Extract: func(i testItem) any { return i.Title }})
	p.AddField(format.Field[testItem]{Name: "state", Header: "STATE", Extract: func(i testItem) any { return i.State }})
	return p
}

// --- Table / TSV output ---

func TestPrinter_TTY_TableHasHeader(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, true, format.OutputConfig{})
	p.AddItem(testItem{1, "Fix auth", "OPEN"})
	p.AddItem(testItem{2, "Bump deps", "MERGED"})
	require.NoError(t, p.Render())
	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "TITLE")
	assert.Contains(t, out, "STATE")
	assert.Contains(t, out, "Fix auth")
	assert.Contains(t, out, "Bump deps")
}

func TestPrinter_NonTTY_TSVNoHeader(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{})
	p.AddItem(testItem{1, "Fix auth", "OPEN"})
	require.NoError(t, p.Render())
	out := buf.String()
	assert.NotContains(t, out, "ID")
	assert.NotContains(t, out, "TITLE")
	assert.Contains(t, out, "Fix auth")
	// tab-separated
	assert.Contains(t, out, "\t")
}

func TestPrinter_Empty_NoOutput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, true, format.OutputConfig{})
	require.NoError(t, p.Render())
	assert.Empty(t, buf.String())
}

// --- JSON output ---

func TestPrinter_JSON_Array_AllFields(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{Format: format.FormatJSON})
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	p.AddItem(testItem{43, "Bump deps", "OPEN"})
	require.NoError(t, p.Render())
	out := strings.TrimSpace(buf.String())
	// must be a JSON array
	assert.True(t, strings.HasPrefix(out, "["), "expected JSON array, got: %s", out)
	assert.True(t, strings.HasSuffix(out, "]"))
	assert.Contains(t, out, `"id":42`)
	assert.Contains(t, out, `"title":"Fix auth"`)
	// All fields ship in OUT2 (field selection deferred)
	assert.Contains(t, out, `"state":"OPEN"`)
}

func TestPrinter_JSON_SingleItem_EmitsObject(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{Format: format.FormatJSON})
	p.SetSingleItem()
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	require.NoError(t, p.Render())
	out := strings.TrimSpace(buf.String())
	assert.True(t, strings.HasPrefix(out, "{"), "single item should emit JSON object, got: %s", out)
	assert.True(t, strings.HasSuffix(out, "}"))
	assert.Contains(t, out, `"id":42`)
}

// --- JQ output ---

func TestPrinter_JQ_WithoutJSON_Error(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{JQExpr: ".[] | .id"})
	p.AddItem(testItem{1, "x", "OPEN"})
	err := p.Render()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--jq requires --json")
}

func TestPrinter_JQ_FilterOutput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{Format: format.FormatJSON, JQExpr: ".[] | .id"})
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	p.AddItem(testItem{43, "Bump deps", "OPEN"})
	require.NoError(t, p.Render())
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, []string{"42", "43"}, lines)
}

func TestPrinter_JQ_InvalidExpression_Error(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{Format: format.FormatJSON, JQExpr: "bad bad bad |||"})
	p.AddItem(testItem{1, "x", "OPEN"})
	err := p.Render()
	require.Error(t, err)
}

// --- YAML output ---

func TestPrinter_YAML_Array(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{Format: format.FormatYAML})
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	p.AddItem(testItem{43, "Bump deps", "MERGED"})
	require.NoError(t, p.Render())
	out := buf.String()
	assert.Contains(t, out, "id: 42")
	assert.Contains(t, out, "title: Fix auth")
	assert.Contains(t, out, "state: OPEN")
	assert.Contains(t, out, "- ")
}

func TestPrinter_YAML_SingleItem(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{Format: format.FormatYAML})
	p.SetSingleItem()
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	require.NoError(t, p.Render())
	out := buf.String()
	assert.Contains(t, out, "id: 42")
	assert.Contains(t, out, "title: Fix auth")
	// not an array
	assert.False(t, strings.HasPrefix(strings.TrimSpace(out), "- "))
}

// --- Template output ---

func TestPrinter_Template_Array(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{
		Format:   format.FormatTemplate,
		Template: `{{range .}}{{.id}}:{{.title}}\n{{end}}`,
	})
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	p.AddItem(testItem{43, "Bump deps", "OPEN"})
	require.NoError(t, p.Render())
	out := buf.String()
	assert.Contains(t, out, "42:Fix auth")
	assert.Contains(t, out, "43:Bump deps")
}

func TestPrinter_Template_SingleItem(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{
		Format:   format.FormatTemplate,
		Template: `{{.id}} - {{.title}}`,
	})
	p.SetSingleItem()
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	require.NoError(t, p.Render())
	assert.Equal(t, "42 - Fix auth", buf.String())
}

func TestPrinter_Template_EmptyExpr_Error(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, format.OutputConfig{Format: format.FormatTemplate})
	p.AddItem(testItem{42, "x", "OPEN"})
	err := p.Render()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--template")
}
