package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// resolvePipelineScheduleClient is the shared preamble for all pipeline-schedule
// handlers: parse hostname + repo, dial backend, type-assert PipelineScheduleClient.
func (h *handlers) resolvePipelineScheduleClient(req mcplib.CallToolRequest) (backend.PipelineScheduleClient, string, string, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return nil, "", "", err
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return nil, "", "", err
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return nil, "", "", err
	}
	sc, err := backend.AsPipelineScheduleClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return sc, ns, slug, nil
}

func (h *handlers) listPipelineSchedules(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	sc, ns, slug, err := h.resolvePipelineScheduleClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	schedules, err := sc.ListPipelineSchedules(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(schedules)
}

func (h *handlers) createPipelineSchedule(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	cron, err := requireString(req, "cron")
	if err != nil {
		return errResultErr(err), nil
	}
	branch, err := requireString(req, "branch")
	if err != nil {
		return errResultErr(err), nil
	}
	// Default enabled to true when not specified.
	// req.GetBool returns false when key absent; we treat absent as true.
	enabled := true
	if req.GetString("enabled", "") != "" {
		enabled = req.GetBool("enabled", true)
	} else {
		// Check if enabled was explicitly set in arguments.
		args, _ := req.Params.Arguments.(map[string]any)
		if _, ok := args["enabled"]; ok {
			enabled = req.GetBool("enabled", true)
		}
	}
	sc, ns, slug, err := h.resolvePipelineScheduleClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	created, err := sc.CreatePipelineSchedule(ns, slug, backend.PipelineScheduleInput{
		CronExpression: cron,
		Branch:         branch,
		Enabled:        enabled,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(created)
}

func (h *handlers) deletePipelineSchedule(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	sc, ns, slug, err := h.resolvePipelineScheduleClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := sc.DeletePipelineSchedule(ns, slug, uuid); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"uuid": uuid, "status": "deleted"})
}
