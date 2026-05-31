package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/reactions"
	"github.com/proggarapsody/bitbottle/pkg/cmd/mcp/argval"
)

func (h *handlers) listPRComments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	cmts, err := client.ListPRComments(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	if req.GetBool("inline_only", false) {
		filtered := make([]backend.PRComment, 0, len(cmts))
		for _, c := range cmts {
			if c.Inline != nil {
				filtered = append(filtered, c)
			}
		}
		cmts = filtered
	}
	if req.GetBool("include_reactions", false) {
		reactor, reactErr := backend.AsCommentReactor(client, hostname)
		if reactErr != nil {
			return errResultErr(reactErr), nil
		}
		var rxnErr error
		cmts, rxnErr = fetchReactionsMCPConcurrent(reactor, project, slug, id, cmts)
		if rxnErr != nil {
			type warningResult struct {
				Comments []backend.PRComment `json:"comments"`
				Warning  string              `json:"warning"`
			}
			return jsonResult(warningResult{Comments: cmts, Warning: fmt.Sprintf("some reactions could not be loaded: %v", rxnErr)})
		}
	}
	return jsonResult(cmts)
}

// fetchReactionsMCPConcurrent fetches reactions for each comment concurrently
// using a bounded worker pool of 4 goroutines. Partial results are returned
// alongside any aggregated error so callers can surface partial data with a warning.
func fetchReactionsMCPConcurrent(reactor backend.CommentReactor, ns, slug string, prID int, cmts []backend.PRComment) ([]backend.PRComment, error) {
	ids := make([]int, len(cmts))
	for i, c := range cmts {
		ids[i] = c.ID
	}
	results, err := reactions.FetchConcurrentByID(ids, func(id int) ([]backend.CommentReaction, error) {
		return reactor.ListCommentReactions(ns, slug, prID, id)
	})
	out := make([]backend.PRComment, len(cmts))
	for i, c := range cmts {
		c.Reactions = results[i]
		out[i] = c
	}
	return out, err
}

func (h *handlers) addPRComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	// MCP-10: the inline anchor is symmetric — inline_path and inline_line
	// must be supplied together. Previously only inline_path without
	// inline_line was caught; inline_line without inline_path slipped through.
	if anchorErr := argval.MutuallyRequired(req.GetArguments(), "inline_path", "inline_line"); anchorErr != nil {
		return errResultArg(anchorErr), nil
	}
	in := backend.AddPRCommentInput{Text: body}
	if inlinePath := req.GetString("inline_path", ""); inlinePath != "" {
		line := req.GetInt("inline_line", 0)
		if line <= 0 {
			return errResult("inline_path requires inline_line (positive integer)"), nil
		}
		side := req.GetString("inline_side", "new")
		if side != "new" && side != "old" {
			return errResult(`inline_side must be "new" or "old"`), nil
		}
		inline := &backend.PRCommentInline{Path: inlinePath, Side: side, Line: line}
		if startLine := req.GetInt("inline_start_line", 0); startLine != 0 {
			if startLine > line {
				return errResult("inline_start_line must be <= inline_line"), nil
			}
			inline.StartLine = startLine
		}
		in.Inline = inline
	}
	if parent := req.GetInt("parent_id", 0); parent > 0 {
		p := parent
		in.Parent = &p
	}
	c, err := client.AddPRComment(project, slug, id, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(c)
}

func (h *handlers) editPRComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	commentID, cidErr := requireIntArg(req, "comment_id")
	if cidErr != nil {
		return cidErr, nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	c, err := client.EditPRComment(project, slug, id, commentID, body)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(c)
}

func (h *handlers) deletePRComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	commentID, cidErr := requireIntArg(req, "comment_id")
	if cidErr != nil {
		return cidErr, nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.DeletePRComment(project, slug, id, commentID); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

// submitPRReview implements the submit_pr_review tool — the MCP-side mirror
// of `pr review`. It takes an action ("approve" / "request_changes" /
// "comment"; default "comment" when body or inline_comments is set), an
// optional top-level body, and an optional inline_comments array of
// {path, line, body, [start_line], [side]} objects. Validation matches the
// CLI: at least one of action/body/inline_comments must be set, and each
// inline entry must carry a non-empty path, a positive line, and a body.
func (h *handlers) submitPRReview(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}

	action := req.GetString("action", "")
	body := req.GetString("body", "")

	inlineRaw, _ := req.GetArguments()["inline_comments"].([]any)
	inline := make([]backend.SubmitReviewInline, 0, len(inlineRaw))
	for i, item := range inlineRaw {
		obj, ok := item.(map[string]any)
		if !ok {
			return errResult(fmt.Sprintf("inline_comments[%d]: must be an object", i)), nil
		}
		ic, perr := parseInlineCommentObject(obj)
		if perr != nil {
			return errResult(fmt.Sprintf("inline_comments[%d]: %v", i, perr)), nil
		}
		inline = append(inline, ic)
	}

	if action == "" {
		if body == "" && len(inline) == 0 {
			return errResult(`one of action ("approve", "request_changes", "comment"), body, or inline_comments is required`), nil
		}
		action = "comment"
	}
	switch action {
	case "approve", "request_changes", "comment":
	default:
		return errResult(fmt.Sprintf(`action must be "approve", "request_changes", or "comment" (got %q)`, action)), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.SubmitReview(project, slug, id, backend.SubmitReviewInput{
		Action: action,
		Body:   body,
		Inline: inline,
	}); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

// parseInlineCommentObject pulls a SubmitReviewInline out of a raw JSON
// object, applying the same shape rules the CLI's parseInlineReviewSpec
// enforces: path/line/body required, side defaults to "new", start_line
// must be <= line when present.
func parseInlineCommentObject(obj map[string]any) (backend.SubmitReviewInline, error) {
	path, _ := obj["path"].(string)
	if path == "" {
		return backend.SubmitReviewInline{}, fmt.Errorf("path is required")
	}
	body, _ := obj["body"].(string)
	if body == "" {
		return backend.SubmitReviewInline{}, fmt.Errorf("body is required")
	}
	lineF, ok := obj["line"].(float64)
	if !ok || lineF <= 0 {
		return backend.SubmitReviewInline{}, fmt.Errorf("line must be a positive integer")
	}
	out := backend.SubmitReviewInline{
		Path: path,
		Line: int(lineF),
		Body: body,
		Side: "new",
	}
	if side, ok := obj["side"].(string); ok && side != "" {
		if side != "new" && side != "old" {
			return backend.SubmitReviewInline{}, fmt.Errorf(`side must be "new" or "old" (got %q)`, side)
		}
		out.Side = side
	}
	if startF, ok := obj["start_line"].(float64); ok && startF > 0 {
		start := int(startF)
		if start > out.Line {
			return backend.SubmitReviewInline{}, fmt.Errorf("start_line %d must be <= line %d", start, out.Line)
		}
		out.StartLine = start
	}
	return out, nil
}

func (h *handlers) resolvePRComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	commentID, cidErr := requireIntArg(req, "comment_id")
	if cidErr != nil {
		return cidErr, nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	resolver, err := backend.AsPRCommentResolver(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := resolver.ResolvePRComment(project, slug, id, commentID); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}
