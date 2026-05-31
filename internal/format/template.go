package format

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"
)

// WriteTemplate parses tmpl as a Go text/template (with the gh-compatible
// FuncMap) and executes it against v, writing the result to w.
func WriteTemplate(w io.Writer, tmpl string, v any) error {
	t, err := template.New("output").Funcs(templateFuncs()).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("--template: parse: %w", err)
	}
	if err := t.Execute(w, v); err != nil {
		return fmt.Errorf("--template: execute: %w", err)
	}
	// Terminate the rendered output with a newline so piped consumers and
	// terminals get a clean line break (parity with --json/--yaml, which the
	// encoders newline-terminate). text/template never appends one itself.
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("--template: execute: %w", err)
	}
	return nil
}

// templateFuncs returns the gh-compatible FuncMap exposed to user templates.
//
// The set mirrors `gh` so existing gh-style templates port over directly:
// color, autocolor, truncate, timeago, pluck, join, tablerender, hyperlink.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"color":       colorFunc,
		"autocolor":   colorFunc, // no TTY context available; safe choice is "always color"
		"truncate":    truncateFunc,
		"timeago":     timeagoFunc,
		"pluck":       pluckFunc,
		"join":        joinFunc,
		"tablerender": tablerenderFunc,
		"hyperlink":   hyperlinkFunc,
	}
}

var ansiColors = map[string]string{
	"bold":    "\x1b[1m",
	"red":     "\x1b[31m",
	"yellow":  "\x1b[33m",
	"green":   "\x1b[32m",
	"gray":    "\x1b[90m",
	"white":   "\x1b[37m",
	"cyan":    "\x1b[36m",
	"blue":    "\x1b[34m",
	"magenta": "\x1b[35m",
}

const ansiReset = "\x1b[0m"

func colorFunc(name, text string) string {
	code, ok := ansiColors[strings.ToLower(name)]
	if !ok {
		return text
	}
	return code + text + ansiReset
}

func truncateFunc(maxWidth int, s string) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth == 1 {
		return "…"
	}
	return string(runes[:maxWidth-1]) + "…"
}

func timeagoFunc(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		m := int(d.Minutes())
		return pluralize(m, "minute") + " ago"
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		return pluralize(h, "hour") + " ago"
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		return pluralize(days, "day") + " ago"
	}
	months := days / 30
	if months < 12 {
		return pluralize(months, "month") + " ago"
	}
	years := days / 365
	return pluralize(years, "year") + " ago"
}

func pluralize(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

func pluckFunc(field string, items []any) []any {
	out := make([]any, 0, len(items))
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			if v, present := m[field]; present {
				out = append(out, v)
			}
		}
	}
	return out
}

func joinFunc(sep string, items []any) string {
	parts := make([]string, len(items))
	for i, it := range items {
		parts[i] = fmt.Sprintf("%v", it)
	}
	return strings.Join(parts, sep)
}

func tablerenderFunc(rows [][]string) string {
	var b strings.Builder
	for _, row := range rows {
		b.WriteString(strings.Join(row, "\t"))
		b.WriteByte('\n')
	}
	return b.String()
}

func hyperlinkFunc(url, text string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}
