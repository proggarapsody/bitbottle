package mcp

import (
	"context"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) listPRTasks(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	stateFilter := req.GetString("state", "open")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	cmts, err := client.ListPRComments(project, slug, prID)
	if err != nil {
		return errResultErr(err), nil
	}

	// The current Cloud adapter exposes ordinary comments without task
	// severity. Keep that fallback tied to capability, not to whether this
	// particular response happens to contain a BLOCKER comment.
	if _, supportsTaskState := client.(backend.PRCommentStateSetter); !supportsTaskState {
		return jsonResult(cmts)
	}
	result := make([]backend.PRComment, 0)
	sf := strings.ToLower(stateFilter)
	for _, c := range cmts {
		if c.Severity != "BLOCKER" {
			continue
		}
		switch sf {
		case "resolved":
			if strings.EqualFold(c.State, "RESOLVED") {
				result = append(result, c)
			}
		case "all":
			result = append(result, c)
		default: // "open"
			if c.State == "" || strings.EqualFold(c.State, "OPEN") {
				result = append(result, c)
			}
		}
	}

	return jsonResult(result)
}

func (h *handlers) createPRTask(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	// Cloud currently ignores AddPRCommentInput.Severity. Check the current
	// adapter's task capability before writing an ordinary comment as a task.
	if _, err := backend.AsPRCommentStateSetter(client, hostname); err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	in := backend.AddPRCommentInput{
		Text:     body,
		Severity: "BLOCKER",
	}
	if parentID := req.GetInt("parent_comment_id", 0); parentID != 0 {
		p := parentID
		in.Parent = &p
	}

	c, err := client.AddPRComment(project, slug, prID, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(struct {
		ID       int    `json:"id"`
		Text     string `json:"text"`
		Severity string `json:"severity"`
		State    string `json:"state"`
	}{ID: c.ID, Text: c.Text, Severity: "BLOCKER", State: "OPEN"})
}

func (h *handlers) resolvePRTask(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	return h.setPRTaskState(req, "RESOLVED")
}

func (h *handlers) reopenPRTask(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	return h.setPRTaskState(req, "OPEN")
}

func (h *handlers) setPRTaskState(req mcplib.CallToolRequest, state string) (*mcplib.CallToolResult, error) {
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
	taskID := req.GetInt("task_id", 0)
	if taskID == 0 {
		return errResult("missing required parameter: task_id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	setter, err := backend.AsPRCommentStateSetter(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	project = strings.ToUpper(project)
	if err := setter.SetPRCommentState(project, slug, prID, taskID, state); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(struct {
		ID    int    `json:"id"`
		State string `json:"state"`
	}{ID: taskID, State: state})
}
