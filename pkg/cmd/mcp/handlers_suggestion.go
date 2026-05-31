package mcp

import (
	"context"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) prSuggestionApply(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	prID := req.GetInt("pr_id", 0)
	if prID == 0 {
		return errResult("missing required parameter: pr_id"), nil
	}
	commentID, cidErr := requireIntArg(req, "comment_id")
	if cidErr != nil {
		return cidErr, nil
	}
	suggestionID := req.GetInt("suggestion_id", 0)
	if suggestionID == 0 {
		return errResult("missing required parameter: suggestion_id"), nil
	}
	preview := req.GetBool("preview", false)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	applier, err := backend.AsSuggestionApplier(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)

	if preview {
		body, err := applier.GetSuggestionPreview(project, slug, prID, commentID)
		if err != nil {
			return errResultErr(err), nil
		}
		return jsonResult(struct {
			Body string `json:"body"`
		}{Body: body})
	}

	result, err := applier.ApplySuggestion(project, slug, prID, commentID, suggestionID)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(struct {
		CommitHash    string `json:"commit_hash"`
		CommitMessage string `json:"commit_message"`
	}{CommitHash: result.CommitHash, CommitMessage: result.CommitMessage})
}
