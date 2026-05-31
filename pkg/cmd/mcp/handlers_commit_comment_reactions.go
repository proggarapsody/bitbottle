package mcp

import (
	"context"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listCommitCommentReactions(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	hash, err := requireString(req, "hash")
	if err != nil {
		return errResultErr(err), nil
	}
	commentID, cidErr := requireIntArg(req, "comment_id")
	if cidErr != nil {
		return cidErr, nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	reactor, err := backend.AsCommitCommentReactor(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	reactions, err := reactor.ListCommitCommentReactions(project, slug, hash, commentID)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(reactions)
}

func (h *handlers) addCommitCommentReaction(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	hash, err := requireString(req, "hash")
	if err != nil {
		return errResultErr(err), nil
	}
	commentID, cidErr := requireIntArg(req, "comment_id")
	if cidErr != nil {
		return cidErr, nil
	}
	emoji, err := requireString(req, "emoji")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	reactor, err := backend.AsCommitCommentReactor(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	if err := reactor.AddCommitCommentReaction(project, slug, hash, commentID, normaliseEmojiMCP(emoji)); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) removeCommitCommentReaction(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	hash, err := requireString(req, "hash")
	if err != nil {
		return errResultErr(err), nil
	}
	commentID, cidErr := requireIntArg(req, "comment_id")
	if cidErr != nil {
		return cidErr, nil
	}
	emoji, err := requireString(req, "emoji")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	reactor, err := backend.AsCommitCommentReactor(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	if err := reactor.RemoveCommitCommentReaction(project, slug, hash, commentID, normaliseEmojiMCP(emoji)); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}
