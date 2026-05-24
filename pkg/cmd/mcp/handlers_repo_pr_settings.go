package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) getRepoPRSettings(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	prc, err := backend.AsRepoPRSettingsClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	settings, err := prc.GetRepoPRSettings(project, repo)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(settings)
}

func (h *handlers) setRepoPRSettings(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}

	args, _ := req.Params.Arguments.(map[string]any)

	in := backend.RepoPRSettingsInput{}
	noop := true

	if _, ok := args["required_approvers"]; ok {
		v := req.GetInt("required_approvers", 0)
		in.RequiredApprovers = &v
		noop = false
	}
	if _, ok := args["required_all_approvers"]; ok {
		v := req.GetBool("required_all_approvers", false)
		in.RequiredAllApprovers = &v
		noop = false
	}
	if _, ok := args["required_all_tasks_complete"]; ok {
		v := req.GetBool("required_all_tasks_complete", false)
		in.RequiredAllTasksComplete = &v
		noop = false
	}
	if _, ok := args["required_successful_builds"]; ok {
		v := req.GetInt("required_successful_builds", 0)
		in.RequiredSuccessfulBuilds = &v
		noop = false
	}
	if ms := req.GetString("merge_strategy", ""); ms != "" {
		in.MergeStrategy = &ms
		noop = false
	}
	if _, ok := args["allowed_strategies"]; ok {
		raw, _ := args["allowed_strategies"].([]any)
		strats := make([]string, 0, len(raw))
		for _, v := range raw {
			if s, ok := v.(string); ok {
				strats = append(strats, s)
			}
		}
		in.AllowedStrategies = &strats
		noop = false
	}

	if noop {
		return errResult("no fields to update; supply at least one of: required_approvers, required_all_approvers, required_all_tasks_complete, required_successful_builds, merge_strategy, allowed_strategies"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	prc, err := backend.AsRepoPRSettingsClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	settings, err := prc.UpdateRepoPRSettings(project, repo, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(settings)
}
