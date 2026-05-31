package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/reactions"
	"github.com/proggarapsody/bitbottle/pkg/cmd/mcp/argval"
)

func (h *handlers) listCommitComments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	limit := req.GetInt("limit", 0)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	cmts, err := client.ListCommitComments(project, slug, hash, limit)
	if err != nil {
		return errResultErr(err), nil
	}

	if req.GetBool("include_reactions", false) {
		reactor, reactErr := backend.AsCommitCommentReactor(client, hostname)
		if reactErr != nil {
			return errResultErr(reactErr), nil
		}
		var rxnErr error
		cmts, rxnErr = fetchCommitReactionsMCPConcurrent(reactor, project, slug, hash, cmts)
		if rxnErr != nil {
			// Surface partial results with a warning field rather than failing.
			type row struct {
				ID        int                       `json:"id"`
				Author    string                    `json:"author"`
				Body      string                    `json:"body"`
				CreatedAt string                    `json:"createdAt"`
				Reactions []backend.CommentReaction `json:"reactions,omitempty"`
			}
			type warningResult struct {
				Comments []row  `json:"comments"`
				Warning  string `json:"warning"`
			}
			rows := make([]row, 0, len(cmts))
			for _, c := range cmts {
				author := c.Author.Slug
				if author == "" {
					author = c.Author.DisplayName
				}
				rows = append(rows, row{
					ID:        c.ID,
					Author:    author,
					Body:      c.Body,
					CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					Reactions: c.Reactions,
				})
			}
			return jsonResult(warningResult{Comments: rows, Warning: fmt.Sprintf("some reactions could not be loaded: %v", rxnErr)})
		}
	}

	type row struct {
		ID        int                       `json:"id"`
		Author    string                    `json:"author"`
		Body      string                    `json:"body"`
		CreatedAt string                    `json:"createdAt"`
		Reactions []backend.CommentReaction `json:"reactions,omitempty"`
	}
	out := make([]row, 0, len(cmts))
	for _, c := range cmts {
		author := c.Author.Slug
		if author == "" {
			author = c.Author.DisplayName
		}
		out = append(out, row{
			ID:        c.ID,
			Author:    author,
			Body:      c.Body,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Reactions: c.Reactions,
		})
	}
	return jsonResult(out)
}

// fetchCommitReactionsMCPConcurrent fetches reactions for each commit comment
// concurrently using a bounded worker pool of 4 goroutines. Partial results are
// returned alongside any aggregated error so callers can surface partial data with a warning.
func fetchCommitReactionsMCPConcurrent(reactor backend.CommitCommentReactor, ns, slug, hash string, cmts []backend.CommitComment) ([]backend.CommitComment, error) {
	ids := make([]int, len(cmts))
	for i, c := range cmts {
		ids[i] = c.ID
	}
	results, err := reactions.FetchConcurrentByID(ids, func(id int) ([]backend.CommentReaction, error) {
		return reactor.ListCommitCommentReactions(ns, slug, hash, id)
	})
	out := make([]backend.CommitComment, len(cmts))
	for i, c := range cmts {
		c.Reactions = results[i]
		out[i] = c
	}
	return out, err
}

func (h *handlers) addCommitComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	// MCP-11: validate hash format client-side (hex, >= 7 chars) so "a" or
	// "NOT_HEX" never reaches the Cloud API as a generic 404.
	hash, hashErr := argval.Hash(req.GetArguments(), "hash", 7)
	if hashErr != nil {
		return errResultArg(hashErr), nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	c, err := client.AddCommitComment(project, slug, hash, backend.AddCommitCommentInput{Body: body})
	if err != nil {
		return errResultErr(err), nil
	}

	author := c.Author.Slug
	if author == "" {
		author = c.Author.DisplayName
	}
	return jsonResult(map[string]any{"id": c.ID, "author": author, "body": c.Body})
}

func (h *handlers) editCommitComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	c, err := client.EditCommitComment(project, slug, hash, commentID, body)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": c.ID})
}

func (h *handlers) deleteCommitComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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

	if err := client.DeleteCommitComment(project, slug, hash, commentID); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"deleted": true})
}
