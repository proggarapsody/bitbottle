package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listSnippetComments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	snippetID, err := requireString(req, "snippet_id")
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 50)
	sc, workspace, err := h.resolveSnippetClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	comments, err := sc.ListSnippetComments(workspace, snippetID, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(comments)
}

func (h *handlers) addSnippetComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	snippetID, err := requireString(req, "snippet_id")
	if err != nil {
		return errResultErr(err), nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	sc, workspace, err := h.resolveSnippetClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	c, err := sc.AddSnippetComment(workspace, snippetID, body)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(c)
}

func (h *handlers) deleteSnippetComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	snippetID, err := requireString(req, "snippet_id")
	if err != nil {
		return errResultErr(err), nil
	}
	commentIDFloat := req.GetInt("comment_id", 0)
	if commentIDFloat == 0 {
		return errResult("comment_id is required"), nil
	}
	sc, workspace, err := h.resolveSnippetClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := sc.DeleteSnippetComment(workspace, snippetID, commentIDFloat); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{
		"workspace":  workspace,
		"snippet_id": snippetID,
		"comment_id": commentIDFloat,
		"deleted":    true,
	})
}

// resolveSnippetClientForComment re-uses the existing resolveSnippetClient but
// maps the snippet_id parameter name used by comment tools. Since resolveSnippetClient
// reads "workspace" (req) which is the same, we can call it directly — the
// snippet_id is fetched separately in each handler above.
var _ = (*backend.SnippetComment)(nil) // import guard
