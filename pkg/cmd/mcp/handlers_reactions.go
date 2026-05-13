package mcp

import (
	"context"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// normaliseEmojiMCP normalises user-provided emoji input to canonical underscore form.
func normaliseEmojiMCP(emoji string) string {
	return backend.NormaliseEmoji(emoji)
}

func (h *handlers) listCommentReactions(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	prID := req.GetInt("pr_id", 0)
	if prID == 0 {
		return errResult("missing required parameter: pr_id"), nil
	}
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResult("missing required parameter: comment_id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	reactor, err := backend.AsCommentReactor(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	reactions, err := reactor.ListCommentReactions(project, repo, prID, commentID)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(reactions)
}

func (h *handlers) addCommentReaction(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	prID := req.GetInt("pr_id", 0)
	if prID == 0 {
		return errResult("missing required parameter: pr_id"), nil
	}
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResult("missing required parameter: comment_id"), nil
	}
	emoji, err := requireString(req, "emoji")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	reactor, err := backend.AsCommentReactor(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	if err := reactor.AddCommentReaction(project, repo, prID, commentID, normaliseEmojiMCP(emoji)); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) removeCommentReaction(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	prID := req.GetInt("pr_id", 0)
	if prID == 0 {
		return errResult("missing required parameter: pr_id"), nil
	}
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResult("missing required parameter: comment_id"), nil
	}
	emoji, err := requireString(req, "emoji")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	reactor, err := backend.AsCommentReactor(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	if err := reactor.RemoveCommentReaction(project, repo, prID, commentID, normaliseEmojiMCP(emoji)); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}
