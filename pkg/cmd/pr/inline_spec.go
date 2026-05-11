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

// parseInlineReviewSpec parses the `pr review --inline` value, which
// differs from `pr comment add --inline` by carrying the comment body in
// the same string: PATH:LINE:BODY (or PATH:START-END:BODY for ranges).
//
// Splitting on the first two colons (rather than the last) means BODY may
// contain colons freely while PATH may not — same trade-off the gh CLI
// makes with `--field NAME:VALUE`. Side defaults to "new"; pass --inline
// the path-with-side variant via separate flags if the old side is needed
// (RV4 does not expose --side; callers stay on the new side which is the
// only documented use case for compound reviews on Cloud).
func parseInlineReviewSpec(spec string) (backend.SubmitReviewInline, error) {
	parts := strings.SplitN(spec, ":", 3)
	if len(parts) < 3 {
		return backend.SubmitReviewInline{}, fmt.Errorf("--inline: expected PATH:LINE:BODY (got %q)", spec)
	}
	path, lineStr, body := parts[0], parts[1], parts[2]
	if path == "" {
		return backend.SubmitReviewInline{}, fmt.Errorf("--inline: path is required (got %q)", spec)
	}
	if lineStr == "" {
		return backend.SubmitReviewInline{}, fmt.Errorf("--inline: line is required (got %q)", spec)
	}
	if body == "" {
		return backend.SubmitReviewInline{}, fmt.Errorf("--inline: body is required (got %q)", spec)
	}

	startStr, endStr, multi := strings.Cut(lineStr, "-")
	end, err := strconv.Atoi(startStr)
	if err != nil || end <= 0 {
		return backend.SubmitReviewInline{}, fmt.Errorf("--inline: line must be a positive integer (got %q)", startStr)
	}
	out := backend.SubmitReviewInline{Path: path, Side: "new", Line: end, Body: body}
	if multi {
		final, err := strconv.Atoi(endStr)
		if err != nil || final <= 0 {
			return backend.SubmitReviewInline{}, fmt.Errorf("--inline: line must be a positive integer (got %q)", endStr)
		}
		if final < end {
			return backend.SubmitReviewInline{}, fmt.Errorf("--inline: start line %d must be <= end line %d", end, final)
		}
		out.StartLine = end
		out.Line = final
	}
	return out, nil
}
