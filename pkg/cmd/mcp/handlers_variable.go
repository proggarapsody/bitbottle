package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/variable/shared"
)

// variableList handles the variable_list MCP tool.
func (h *handlers) variableList(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	scope := req.GetString("scope", "repository")
	envUUID := req.GetString("env_uuid", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	switch scope {
	case "repository":
		pc, err := backend.AsPipelineClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		vars, err := pc.ListPipelineVariables(ns, slug)
		if err != nil {
			return errResultErr(err), nil
		}
		for i := range vars {
			if vars[i].Secured {
				vars[i].Value = ""
			}
		}
		return jsonResult(vars)

	case "workspace":
		wc, err := backend.AsWorkspaceVariableClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		vars, err := wc.ListWorkspaceVariables(ns)
		if err != nil {
			return errResultErr(err), nil
		}
		for i := range vars {
			if vars[i].Secured {
				vars[i].Value = ""
			}
		}
		return jsonResult(vars)

	case "deployment":
		if envUUID == "" {
			return errResult("env_uuid is required for scope=deployment"), nil
		}
		dc, err := backend.AsDeploymentClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		vars, err := dc.ListEnvVariables(ns, slug, envUUID)
		if err != nil {
			return errResultErr(err), nil
		}
		for i := range vars {
			if vars[i].Secured {
				vars[i].Value = ""
			}
		}
		return jsonResult(vars)

	default:
		return errResult(fmt.Sprintf("unknown scope %q; valid: repository, workspace, deployment", scope)), nil
	}
}

// variableView handles the variable_view MCP tool.
func (h *handlers) variableView(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	scope := req.GetString("scope", "repository")
	envUUID := req.GetString("env_uuid", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	ops, err := shared.ResolveVariableOps(scope, client, hostname, ns, slug, envUUID)
	if err != nil {
		return errResultErr(err), nil
	}
	v, err := ops.GetVariableByKey(key)
	if err != nil {
		return errResultErr(err), nil
	}
	if v.Secured {
		v.Value = ""
	}
	return jsonResult(v)
}

// variableSet handles the variable_set MCP tool.
func (h *handlers) variableSet(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	value, err := requireString(req, "value")
	if err != nil {
		return errResultErr(err), nil
	}
	secured := req.GetBool("secured", false)
	scope := req.GetString("scope", "repository")
	envUUID := req.GetString("env_uuid", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	switch scope {
	case "repository":
		pc, err := backend.AsPipelineClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		v, err := pc.SetPipelineVariable(ns, slug, backend.PipelineVariableInput{
			Key:     key,
			Value:   value,
			Secured: secured,
		})
		if err != nil {
			return errResultErr(err), nil
		}
		if v.Secured {
			v.Value = ""
		}
		return jsonResult(v)

	case "workspace":
		wc, err := backend.AsWorkspaceVariableClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		v, err := wc.SetWorkspaceVariable(ns, backend.PipelineVariableInput{
			Key:     key,
			Value:   value,
			Secured: secured,
		})
		if err != nil {
			return errResultErr(err), nil
		}
		if v.Secured {
			v.Value = ""
		}
		return jsonResult(v)

	case "deployment":
		if envUUID == "" {
			return errResult("env_uuid is required for scope=deployment"), nil
		}
		dc, err := backend.AsDeploymentClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		v, err := dc.SetEnvVariable(ns, slug, envUUID, backend.EnvVariableInput{
			Key:     key,
			Value:   value,
			Secured: secured,
		})
		if err != nil {
			return errResultErr(err), nil
		}
		if v.Secured {
			v.Value = ""
		}
		return jsonResult(v)

	default:
		return errResult(fmt.Sprintf("unknown scope %q; valid: repository, workspace, deployment", scope)), nil
	}
}

// variableDelete handles the variable_delete MCP tool.
func (h *handlers) variableDelete(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	repo, err := requireString(req, "repo")
	if err != nil {
		return errResultErr(err), nil
	}
	ns, slug, err := splitRepo(repo)
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	scope := req.GetString("scope", "repository")
	envUUID := req.GetString("env_uuid", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	switch scope {
	case "repository":
		pc, err := backend.AsPipelineClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		if err := pc.DeletePipelineVariable(ns, slug, key); err != nil {
			return errResultErr(err), nil
		}
		return jsonResult(map[string]string{"key": key, "status": "deleted"})

	case "workspace":
		wc, err := backend.AsWorkspaceVariableClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		if err := wc.DeleteWorkspaceVariable(ns, key); err != nil {
			return errResultErr(err), nil
		}
		return jsonResult(map[string]string{"key": key, "status": "deleted"})

	case "deployment":
		if envUUID == "" {
			return errResult("env_uuid is required for scope=deployment"), nil
		}
		dc, err := backend.AsDeploymentClient(client, hostname)
		if err != nil {
			return errResultErr(err), nil
		}
		// Find by key, delete by UUID.
		vars, err := dc.ListEnvVariables(ns, slug, envUUID)
		if err != nil {
			return errResultErr(err), nil
		}
		var varUUID string
		for _, v := range vars {
			if v.Key == key {
				varUUID = v.UUID
				break
			}
		}
		if varUUID == "" {
			return errResult(fmt.Sprintf("variable %q not found in environment %s", key, envUUID)), nil
		}
		if err := dc.DeleteEnvVariable(ns, slug, envUUID, varUUID); err != nil {
			return errResultErr(err), nil
		}
		return jsonResult(map[string]string{"key": key, "status": "deleted"})

	default:
		return errResult(fmt.Sprintf("unknown scope %q; valid: repository, workspace, deployment", scope)), nil
	}
}
