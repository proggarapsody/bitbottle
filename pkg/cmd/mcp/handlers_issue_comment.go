package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

func (h *handlers) listIssueComments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	comments, err := ic.ListIssueComments(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(comments)
}

func (h *handlers) addIssueComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	comment, err := ic.AddIssueComment(project, slug, id, body)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(comment)
}

func (h *handlers) editIssueComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: comment_id")), nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	comment, err := ic.EditIssueComment(project, slug, id, commentID, body)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(comment)
}

func (h *handlers) deleteIssueComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: comment_id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ic.DeleteIssueComment(project, slug, id, commentID); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"deleted": true, "comment_id": commentID})
}
