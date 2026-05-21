package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveSnippetClient is the small adapter every snippet handler shares:
// pick the host, dial the backend, type-assert SnippetClient, and gather
// the workspace arg.
func (h *handlers) resolveSnippetClient(req mcplib.CallToolRequest) (backend.SnippetClient, string, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return nil, "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", err
	}
	sc, err := backend.AsSnippetClient(client, hostname)
	if err != nil {
		return nil, "", err
	}
	return sc, workspace, nil
}

func (h *handlers) listSnippets(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	limit := req.GetInt("limit", 30)
	sc, workspace, err := h.resolveSnippetClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	snippets, err := sc.ListSnippets(workspace, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(snippets)
}

func (h *handlers) viewSnippet(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id, err := requireString(req, "id")
	if err != nil {
		return errResultErr(err), nil
	}
	sc, workspace, err := h.resolveSnippetClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	s, err := sc.GetSnippet(workspace, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(s)
}

func (h *handlers) createSnippet(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	title, err := requireString(req, "title")
	if err != nil {
		return errResultErr(err), nil
	}
	private := req.GetBool("private", false)
	filesJSON := req.GetString("files", "")

	var snippetFiles []backend.SnippetFile
	if filesJSON != "" {
		var raw map[string]string
		if err := json.Unmarshal([]byte(filesJSON), &raw); err != nil {
			return errResult(fmt.Sprintf("invalid files JSON: %v", err)), nil
		}
		for name, content := range raw {
			snippetFiles = append(snippetFiles, backend.SnippetFile{Name: name, Content: content})
		}
	}

	sc, workspace, err := h.resolveSnippetClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	s, err := sc.CreateSnippet(workspace, backend.CreateSnippetInput{
		Title:     title,
		IsPrivate: private,
		Files:     snippetFiles,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(s)
}

func (h *handlers) deleteSnippet(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id, err := requireString(req, "id")
	if err != nil {
		return errResultErr(err), nil
	}
	sc, workspace, err := h.resolveSnippetClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := sc.DeleteSnippet(workspace, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"workspace": workspace, "id": id, "deleted": true})
}
