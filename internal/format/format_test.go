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

func newPrinter(w *bytes.Buffer, isTTY bool, jsonFields, jqExpr string) *format.Printer[testItem] {
	p := format.New[testItem](w, isTTY, jsonFields, jqExpr)
	p.AddField(format.Field[testItem]{Name: "id", Header: "ID", Extract: func(i testItem) any { return i.ID }})
	p.AddField(format.Field[testItem]{Name: "title", Header: "TITLE", Extract: func(i testItem) any { return i.Title }})
	p.AddField(format.Field[testItem]{Name: "state", Header: "STATE", Extract: func(i testItem) any { return i.State }})
	return p
}

// --- Table / TSV output ---

func TestPrinter_TTY_TableHasHeader(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, true, "", "")
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
	p := newPrinter(&buf, false, "", "")
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
	p := newPrinter(&buf, true, "", "")
	require.NoError(t, p.Render())
	assert.Empty(t, buf.String())
}

// --- JSON output ---

func TestPrinter_JSON_Array_SelectedFields(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, "id,title", "")
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	p.AddItem(testItem{43, "Bump deps", "OPEN"})
	require.NoError(t, p.Render())
	out := strings.TrimSpace(buf.String())
	// must be a JSON array
	assert.True(t, strings.HasPrefix(out, "["), "expected JSON array, got: %s", out)
	assert.True(t, strings.HasSuffix(out, "]"))
	assert.Contains(t, out, `"id":42`)
	assert.Contains(t, out, `"title":"Fix auth"`)
	// state not requested
	assert.NotContains(t, out, `"state"`)
}

func TestPrinter_JSON_SingleItem_EmitsObject(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, "id,title", "")
	p.SetSingleItem()
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	require.NoError(t, p.Render())
	out := strings.TrimSpace(buf.String())
	assert.True(t, strings.HasPrefix(out, "{"), "single item should emit JSON object, got: %s", out)
	assert.True(t, strings.HasSuffix(out, "}"))
	assert.Contains(t, out, `"id":42`)
}

func TestPrinter_JSON_UnknownField_Error(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, "bad", "")
	p.AddItem(testItem{1, "x", "OPEN"})
	err := p.Render()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown field "bad"`)
	assert.Contains(t, err.Error(), "id")
	assert.Contains(t, err.Error(), "title")
}

// --- JQ output ---

func TestPrinter_JQ_WithoutJSON_Error(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, "", ".[] | .id")
	p.AddItem(testItem{1, "x", "OPEN"})
	err := p.Render()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--jq requires --json")
}

func TestPrinter_JQ_FilterOutput(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, "id,state", ".[] | .id")
	p.AddItem(testItem{42, "Fix auth", "OPEN"})
	p.AddItem(testItem{43, "Bump deps", "OPEN"})
	require.NoError(t, p.Render())
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, []string{"42", "43"}, lines)
}

func TestPrinter_JQ_InvalidExpression_Error(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	p := newPrinter(&buf, false, "id", "bad bad bad |||")
	p.AddItem(testItem{1, "x", "OPEN"})
	err := p.Render()
	require.Error(t, err)
}
