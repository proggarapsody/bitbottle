package mcp

import (
	"context"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func (h *handlers) triggerPipeline(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	branch, err := requireString(req, "branch")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	tc, err := backend.AsPipelineTriggerClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	vars, err := parseMCPVariables(req.GetString("variables", ""))
	if err != nil {
		return errResultErr(err), nil
	}

	result, err := tc.TriggerPipeline(ns, slug, backend.PipelineTriggerInput{
		Branch:    branch,
		Variables: vars,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(result)
}

// parseMCPVariables converts a comma-separated "key=value,key=value" string
// into a PipelineVariable slice. An empty string returns nil, nil.
func parseMCPVariables(raw string) ([]backend.PipelineVariable, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	vars := make([]backend.PipelineVariable, 0, len(parts))
	for _, kv := range parts {
		kv = strings.TrimSpace(kv)
		if kv == "" {
			continue
		}
		idx := strings.IndexByte(kv, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid variable %q: expected key=value format", kv)
		}
		vars = append(vars, backend.PipelineVariable{
			Key:   kv[:idx],
			Value: kv[idx+1:],
		})
	}
	return vars, nil
}
