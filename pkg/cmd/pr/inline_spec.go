package pr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// parseInlineSpec parses a "path:line" or "path:start-end" anchor spec used
// by --inline. The path may itself contain colons; only the final colon
// separates path from line. The side argument selects which diff side the
// anchor lives on ("new" or "old"); empty defaults to "new".
//
// Errors are descriptive and meant to surface directly to the user via the
// CLI flag handler, so wrapping them is unnecessary at the call site.
func parseInlineSpec(spec, side string) (*backend.PRCommentInline, error) {
	switch side {
	case "", "new":
		side = "new"
	case "old":
		// ok
	default:
		return nil, fmt.Errorf("--side must be \"new\" or \"old\" (got %q)", side)
	}

	idx := strings.LastIndex(spec, ":")
	if idx < 0 {
		return nil, fmt.Errorf("--inline: expected path:line (got %q)", spec)
	}
	path := spec[:idx]
	rest := spec[idx+1:]
	if path == "" {
		return nil, fmt.Errorf("--inline: path is required (got %q)", spec)
	}
	if rest == "" {
		return nil, fmt.Errorf("--inline: line is required (got %q)", spec)
	}

	startStr, endStr, multi := strings.Cut(rest, "-")
	end, err := strconv.Atoi(startStr)
	if err != nil || end <= 0 {
		return nil, fmt.Errorf("--inline: line must be a positive integer (got %q)", startStr)
	}
	out := &backend.PRCommentInline{Path: path, Side: side, Line: end}
	if multi {
		final, err := strconv.Atoi(endStr)
		if err != nil || final <= 0 {
			return nil, fmt.Errorf("--inline: line must be a positive integer (got %q)", endStr)
		}
		if final < end {
			return nil, fmt.Errorf("--inline: start line %d must be <= end line %d", end, final)
		}
		out.StartLine = end
		out.Line = final
	}
	return out, nil
}
