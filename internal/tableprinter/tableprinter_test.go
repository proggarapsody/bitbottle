package tableprinter_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/tableprinter"
)

func TestTablePrinter_Render_Empty(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, false, 80)
	require.NoError(t, tp.Render())
	assert.Empty(t, buf.String())
}

func TestTablePrinter_Render_SingleRow_NonTTY(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, false, 80)
	tp.AddField("hello")
	tp.AddField("world")
	tp.EndRow()
	require.NoError(t, tp.Render())
	assert.Equal(t, "hello\tworld\n", buf.String())
}

func TestTablePrinter_Render_MultiRow_NonTTY(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, false, 80)
	tp.AddField("a")
	tp.AddField("b")
	tp.EndRow()
	tp.AddField("cc")
	tp.AddField("dd")
	tp.EndRow()
	require.NoError(t, tp.Render())
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	assert.Len(t, lines, 2)
	assert.Equal(t, "a\tb", lines[0])
	assert.Equal(t, "cc\tdd", lines[1])
}

func TestTablePrinter_Render_TTY_Aligned(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, true, 80)
	tp.AddField("short")
	tp.AddField("col2")
	tp.EndRow()
	tp.AddField("a-much-longer-field")
	tp.AddField("col2-val")
	tp.EndRow()
	require.NoError(t, tp.Render())

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	require.Len(t, lines, 2)
	// In TTY mode, first column is padded to the longest value.
	assert.True(t, strings.HasPrefix(lines[0], "short              "), "first col should be padded: %q", lines[0])
	assert.True(t, strings.HasPrefix(lines[1], "a-much-longer-field"), "second row first col should not be padded (it IS the longest): %q", lines[1])
}

func TestTablePrinter_Render_SingleColumn(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, false, 80)
	tp.AddField("only")
	tp.EndRow()
	require.NoError(t, tp.Render())
	assert.Equal(t, "only\n", buf.String())
}

func TestTablePrinter_Render_EmptyField(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, false, 80)
	tp.AddField("")
	tp.AddField("val")
	tp.EndRow()
	require.NoError(t, tp.Render())
	assert.Equal(t, "\tval\n", buf.String())
}

// errWriter always returns an error.
type errWriter struct{}

func (e *errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestTablePrinter_Render_PropagatesWriteError(t *testing.T) {
	t.Parallel()
	tp := tableprinter.New(&errWriter{}, false, 80)
	tp.AddField("x")
	tp.EndRow()
	err := tp.Render()
	require.Error(t, err)
}

// TestTablePrinter_Render_TTY_HeaderPrinted verifies that AddHeader outputs a
// header row before data rows in TTY mode and includes the header in column
// width measurements.
func TestTablePrinter_Render_TTY_HeaderPrinted(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, true, 200)
	tp.AddHeader("NAME", "PROJECT")
	tp.AddField("svc")
	tp.AddField("PROJ")
	tp.EndRow()
	require.NoError(t, tp.Render())

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	require.Len(t, lines, 2)
	assert.True(t, strings.HasPrefix(lines[0], "NAME"), "first line should be header")
	assert.True(t, strings.HasPrefix(lines[1], "svc "), "data row should be padded to header width")
}

// TestTablePrinter_Render_NonTTY_HeaderSuppressed verifies that in non-TTY mode
// the header is not emitted so piped output stays machine-readable.
func TestTablePrinter_Render_NonTTY_HeaderSuppressed(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, false, 80)
	tp.AddHeader("NAME", "PROJECT")
	tp.AddField("svc")
	tp.AddField("PROJ")
	tp.EndRow()
	require.NoError(t, tp.Render())

	assert.Equal(t, "svc\tPROJ\n", buf.String())
}

// TestTablePrinter_Render_TTY_MaxWidthTruncates verifies that the last column is
// truncated with an ellipsis when its content would exceed maxWidth.
func TestTablePrinter_Render_TTY_MaxWidthTruncates(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	// col1 = 5 chars padded + 1 tab = 6 used; maxWidth = 10 → last col gets 4 runes.
	tp := tableprinter.New(&buf, true, 10)
	tp.AddField("hello") // padded to 5 (only col; widths[0]=5)
	tp.AddField("this-is-very-long")
	tp.EndRow()
	require.NoError(t, tp.Render())

	line := strings.TrimSuffix(buf.String(), "\n")
	parts := strings.SplitN(line, "\t", 2)
	require.Len(t, parts, 2)
	// last part should be truncated; original is longer than remaining space
	assert.True(t, strings.HasSuffix(parts[1], "…"), "last col should end with ellipsis: %q", parts[1])
}

// TestTablePrinter_Render_UTF8_ColumnWidth verifies that multi-byte UTF-8
// characters are measured by rune count, not byte length, for alignment.
func TestTablePrinter_Render_UTF8_ColumnWidth(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, true, 200)
	// "café" is 4 runes but 5 bytes — alignment should use rune count.
	tp.AddField("café")
	tp.AddField("x")
	tp.EndRow()
	tp.AddField("ab")
	tp.AddField("y")
	tp.EndRow()
	require.NoError(t, tp.Render())

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	require.Len(t, lines, 2)
	// "café" is 4 runes; "ab" is 2 runes → first col padded to 4 runes.
	// "ab" row first col should be "ab  " (padded to 4 rune-width).
	assert.True(t, strings.HasPrefix(lines[1], "ab  "), "UTF-8 col should be padded by rune count: %q", lines[1])
}

// TestTablePrinter_Render_TTY_ANSIColorDoesNotBreakAlignment verifies that
// ANSI escape sequences are excluded from column-width measurements so that
// colored cells still align with plain-text cells.
func TestTablePrinter_Render_TTY_ANSIColorDoesNotBreakAlignment(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	tp := tableprinter.New(&buf, true, 200)
	// row 1: "SUCCESSFUL" wrapped in green ANSI codes (10 visible chars)
	tp.AddField("\033[32mSUCCESSFUL\033[0m")
	tp.AddField("build-1")
	tp.EndRow()
	// row 2: "FAILED" plain (6 visible chars) — should be padded to 10
	tp.AddField("FAILED")
	tp.AddField("lint-1")
	tp.EndRow()
	require.NoError(t, tp.Render())

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	require.Len(t, lines, 2)

	// Split on tab to get columns
	parts0 := strings.SplitN(lines[0], "\t", 2)
	parts1 := strings.SplitN(lines[1], "\t", 2)
	require.Len(t, parts0, 2)
	require.Len(t, parts1, 2)

	// The visible content of col0 in both rows should be padded to the same visual width.
	// "FAILED" (6 chars) must be padded with 4 spaces to match "SUCCESSFUL" (10 visible chars).
	assert.Equal(t, "FAILED    ", parts1[0], "FAILED should be padded to 10 visible chars")
}

func BenchmarkTablePrinter_Render_1000Rows(b *testing.B) {
	for range b.N {
		tp := tableprinter.New(io.Discard, false, 80)
		for range 1000 {
			tp.AddField("my-service")
			tp.AddField("MYPROJ")
			tp.AddField("git")
			tp.EndRow()
		}
		_ = tp.Render()
	}
}
