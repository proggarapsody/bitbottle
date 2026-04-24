package tableprinter

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/aleksey/bitbottle/internal/text"
)

// TablePrinter renders aligned tabular output.
type TablePrinter struct {
	w        io.Writer
	isTTY    bool
	maxWidth int
	header   []string
	rows     [][]string
	cur      []string
}

func New(w io.Writer, isTTY bool, maxWidth int) *TablePrinter {
	return &TablePrinter{w: w, isTTY: isTTY, maxWidth: maxWidth}
}

// AddHeader sets column headers. In TTY mode they are printed before data rows;
// in non-TTY mode they are suppressed so piped output stays machine-readable.
func (t *TablePrinter) AddHeader(fields ...string) {
	t.header = fields
}

func (t *TablePrinter) AddField(s string) {
	t.cur = append(t.cur, s)
}

func (t *TablePrinter) EndRow() {
	t.rows = append(t.rows, t.cur)
	t.cur = nil
}

// Render writes all rows. In TTY mode columns are padded to align and the last
// column is truncated to respect maxWidth; in non-TTY mode fields are
// tab-separated and headers are omitted.
func (t *TablePrinter) Render() error {
	if len(t.rows) == 0 {
		return nil
	}

	// In TTY mode, include the header in column-width measurements.
	measureRows := t.rows
	if t.isTTY && len(t.header) > 0 {
		measureRows = append([][]string{t.header}, t.rows...)
	}

	// Compute max column widths using rune counts (correct for multi-byte UTF-8).
	cols := 0
	for _, row := range measureRows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	widths := make([]int, cols)
	if t.isTTY {
		for _, row := range measureRows {
			for i, cell := range row {
				if w := utf8.RuneCountInString(cell); w > widths[i] {
					widths[i] = w
				}
			}
		}
	}

	renderRow := func(row []string) error {
		parts := make([]string, len(row))
		for i, cell := range row {
			switch {
			case t.isTTY && i < len(row)-1:
				// Intermediate columns: pad to column width for alignment.
				parts[i] = fmt.Sprintf("%-*s", widths[i], cell)
			case t.isTTY && t.maxWidth > 0:
				// Last column in TTY mode: truncate to remaining terminal width.
				used := 0
				for j := 0; j < len(row)-1; j++ {
					used += widths[j] + 1 // padded width + tab separator
				}
				if remaining := t.maxWidth - used; remaining > 0 {
					parts[i] = text.Truncate(cell, remaining)
				} else {
					parts[i] = cell
				}
			default:
				parts[i] = cell
			}
		}
		_, err := fmt.Fprintln(t.w, strings.Join(parts, "\t"))
		return err
	}

	// Print header first in TTY mode (suppressed for non-TTY / piped output).
	if t.isTTY && len(t.header) > 0 {
		if err := renderRow(t.header); err != nil {
			return err
		}
	}

	for _, row := range t.rows {
		if err := renderRow(row); err != nil {
			return err
		}
	}
	return nil
}
