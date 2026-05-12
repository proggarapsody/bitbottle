package mcp

import (
	"context"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// splitRepo splits a "WORKSPACE/REPO" string into (namespace, slug).
// Returns an error when the string does not contain exactly one slash separator
// or either part is empty.
func splitRepo(repo string) (string, string, error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("repo must be in WORKSPACE/REPO format, got %q", repo)
	}
	return parts[0], parts[1], nil
}

// resolveDeploymentClient is the shared preamble for all deployment handlers:
// parse hostname + repo, dial backend, type-assert DeploymentClient.
func (h *handlers) resolveDeploymentClient(req mcplib.CallToolRequest) (backend.DeploymentClient, string, string, error) {
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
	dc, err := backend.AsDeploymentClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return dc, ns, slug, nil
}

func (h *handlers) listDeployments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	limit := req.GetInt("limit", 10)
	dc, ns, slug, err := h.resolveDeploymentClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	deployments, err := dc.ListDeployments(ns, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(deployments)
}

func (h *handlers) getDeployment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	dc, ns, slug, err := h.resolveDeploymentClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	d, err := dc.GetDeployment(ns, slug, uuid)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(d)
}

func (h *handlers) listEnvironments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	dc, ns, slug, err := h.resolveDeploymentClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	envs, err := dc.ListEnvironments(ns, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(envs)
}

func (h *handlers) createEnvironment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	envType, err := requireString(req, "type")
	if err != nil {
		return errResultErr(err), nil
	}
	rank := req.GetInt("rank", 0)
	dc, ns, slug, err := h.resolveDeploymentClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	env, err := dc.CreateEnvironment(ns, slug, backend.CreateEnvironmentInput{
		Name: name,
		Type: envType,
		Rank: rank,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(env)
}

func (h *handlers) deleteEnvironment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	envUUID, err := requireString(req, "env_uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	dc, ns, slug, err := h.resolveDeploymentClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := dc.DeleteEnvironment(ns, slug, envUUID); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"uuid": envUUID, "status": "deleted"})
}

func (h *handlers) listEnvVariables(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	envUUID, err := requireString(req, "env_uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	dc, ns, slug, err := h.resolveDeploymentClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	vars, err := dc.ListEnvVariables(ns, slug, envUUID)
	if err != nil {
		return errResultErr(err), nil
	}
	// Defensive: blank secured values before serialising.
	for i := range vars {
		if vars[i].Secured {
			vars[i].Value = ""
		}
	}
	return jsonResult(vars)
}

func (h *handlers) setEnvVariable(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	envUUID, err := requireString(req, "env_uuid")
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
	dc, ns, slug, err := h.resolveDeploymentClient(req)
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
}

func (h *handlers) deleteEnvVariable(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	envUUID, err := requireString(req, "env_uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	varUUID, err := requireString(req, "key")
	if err != nil {
		return errResultErr(err), nil
	}
	dc, ns, slug, err := h.resolveDeploymentClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := dc.DeleteEnvVariable(ns, slug, envUUID, varUUID); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"uuid": varUUID, "status": "deleted"})
}
