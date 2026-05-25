package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolveIssueAttacherClient is the small adapter every issue-attachment
// handler shares: pick the host, dial the backend, type-assert IssueAttacher,
// and gather the project/slug args.
func (h *handlers) resolveIssueAttacherClient(req mcplib.CallToolRequest) (backend.IssueAttacher, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	ia, err := backend.AsIssueAttacher(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return ia, project, slug, nil
}

// resolveIssueVoterClient resolves an IssueVoter client from the request.
func (h *handlers) resolveIssueVoterClient(req mcplib.CallToolRequest) (backend.IssueVoter, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	iv, err := backend.AsIssueVoter(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return iv, project, slug, nil
}

// resolveIssueWatcherClient resolves an IssueWatcher client from the request.
func (h *handlers) resolveIssueWatcherClient(req mcplib.CallToolRequest) (backend.IssueWatcher, string, string, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return nil, "", "", err
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	iw, err := backend.AsIssueWatcher(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return iw, project, slug, nil
}

func (h *handlers) listIssueAttachments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ia, project, slug, err := h.resolveIssueAttacherClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	attachments, err := ia.ListIssueAttachments(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(attachments)
}

func (h *handlers) deleteIssueAttachment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	filename, err := requireString(req, "filename")
	if err != nil {
		return errResultErr(err), nil
	}
	ia, project, slug, err := h.resolveIssueAttacherClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ia.DeleteIssueAttachment(project, slug, id, filename); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "filename": filename, "deleted": true})
}

func (h *handlers) voteIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	iv, project, slug, err := h.resolveIssueVoterClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := iv.VoteIssue(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "voted": true})
}

func (h *handlers) unvoteIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	iv, project, slug, err := h.resolveIssueVoterClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := iv.UnvoteIssue(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "voted": false})
}

func (h *handlers) watchIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	iw, project, slug, err := h.resolveIssueWatcherClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := iw.WatchIssue(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "watching": true})
}

func (h *handlers) unwatchIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	iw, project, slug, err := h.resolveIssueWatcherClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := iw.UnwatchIssue(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "watching": false})
}
