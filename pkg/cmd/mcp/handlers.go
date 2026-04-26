package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

type handlers struct {
	f *factory.Factory
}

func newHandlers(f *factory.Factory) *handlers {
	return &handlers{f: f}
}

func (h *handlers) resolveBackend(hostname string) (backend.Client, error) {
	if hostname != "" {
		return h.f.Backend(hostname)
	}
	cfg, err := h.f.Config()
	if err != nil {
		return nil, err
	}
	hosts := cfg.Hosts()
	switch len(hosts) {
	case 0:
		return nil, fmt.Errorf("not authenticated; run `bitbottle auth login` first")
	case 1:
		return h.f.Backend(hosts[0])
	default:
		return nil, fmt.Errorf("multiple hosts configured; specify hostname")
	}
}

func jsonResult(v any) (*mcplib.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("serialize: %v", err)), nil
	}
	return mcplib.NewToolResultText(string(data)), nil
}

func errResult(msg string) *mcplib.CallToolResult {
	return mcplib.NewToolResultError(msg)
}

// splitTrimmed splits s by sep and trims whitespace from each part.
func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

func requireString(req mcplib.CallToolRequest, key string) (string, error) {
	v := req.GetString(key, "")
	if v == "" {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	return v, nil
}

func (h *handlers) listHosts(_ context.Context, _ mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	cfg, err := h.f.Config()
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(cfg.Hosts())
}

func (h *handlers) listRepos(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	repos, err := client.ListRepos(limit)
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(repos)
}

func (h *handlers) getRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	repo, err := client.GetRepo(project, slug)
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(repo)
}

func (h *handlers) createRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResult(err.Error()), nil
	}
	description := req.GetString("description", "")
	private := req.GetBool("private", false)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	repo, err := client.CreateRepo(project, backend.CreateRepoInput{
		Name:        name,
		SCM:         "git",
		Public:      !private,
		Description: description,
	})
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(repo)
}

func (h *handlers) deleteRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	if err := client.DeleteRepo(project, slug); err != nil {
		return errResult(err.Error()), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) listPRs(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	state := req.GetString("state", "OPEN")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	prs, err := client.ListPRs(project, slug, state, limit)
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(prs)
}

func (h *handlers) getPR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pr, err := client.GetPR(project, slug, id)
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(pr)
}

func (h *handlers) createPR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	title, err := requireString(req, "title")
	if err != nil {
		return errResult(err.Error()), nil
	}
	fromBranch, err := requireString(req, "from_branch")
	if err != nil {
		return errResult(err.Error()), nil
	}
	toBranch, err := requireString(req, "to_branch")
	if err != nil {
		return errResult(err.Error()), nil
	}
	body := req.GetString("body", "")
	draft := req.GetBool("draft", false)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pr, err := client.CreatePR(project, slug, backend.CreatePRInput{
		Title:       title,
		Description: body,
		Draft:       draft,
		FromBranch:  fromBranch,
		ToBranch:    toBranch,
	})
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(pr)
}

func (h *handlers) mergePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}
	strategy := req.GetString("strategy", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pr, err := client.MergePR(project, slug, id, backend.MergePRInput{Strategy: strategy})
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(pr)
}

func (h *handlers) approvePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	if err := client.ApprovePR(project, slug, id); err != nil {
		return errResult(err.Error()), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) getPRDiff(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	diff, err := client.GetPRDiff(project, slug, id)
	if err != nil {
		return errResult(err.Error()), nil
	}
	return mcplib.NewToolResultText(diff), nil
}

func (h *handlers) deleteBranch(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	branch, err := requireString(req, "branch")
	if err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	if err := client.DeleteBranch(project, slug, branch); err != nil {
		return errResult(err.Error()), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) getCurrentUser(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	user, err := client.GetCurrentUser()
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(user)
}

func (h *handlers) listBranches(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	branches, err := client.ListBranches(project, slug, limit)
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(branches)
}

func (h *handlers) listPipelines(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 20)

	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pc, err := backend.AsPipelineClient(client)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pipelines, err := pc.ListPipelines(project, slug, limit)
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(pipelines)
}

func (h *handlers) getPipeline(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pc, err := backend.AsPipelineClient(client)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pl, err := pc.GetPipeline(project, slug, uuid)
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(pl)
}

func (h *handlers) updatePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}
	title := req.GetString("title", "")
	body := req.GetString("body", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pr, err := client.UpdatePR(project, slug, id, backend.UpdatePRInput{
		Title:       title,
		Description: body,
	})
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(pr)
}

func (h *handlers) declinePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	if err := client.DeclinePR(project, slug, id); err != nil {
		return errResult(err.Error()), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) unapprovePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	if err := client.UnapprovePR(project, slug, id); err != nil {
		return errResult(err.Error()), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) readyPR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	if err := client.ReadyPR(project, slug, id); err != nil {
		return errResult(err.Error()), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) requestReview(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}
	reviewers, err := requireString(req, "reviewers")
	if err != nil {
		return errResult(err.Error()), nil
	}

	var users []string
	for _, u := range splitTrimmed(reviewers, ",") {
		if u != "" {
			users = append(users, u)
		}
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	if err := client.RequestReview(project, slug, id, users); err != nil {
		return errResult(err.Error()), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) runPipeline(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResult(err.Error()), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResult(err.Error()), nil
	}
	branch, err := requireString(req, "branch")
	if err != nil {
		return errResult(err.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pc, err := backend.AsPipelineClient(client)
	if err != nil {
		return errResult(err.Error()), nil
	}
	pl, err := pc.RunPipeline(project, slug, backend.RunPipelineInput{Branch: branch})
	if err != nil {
		return errResult(err.Error()), nil
	}
	return jsonResult(pl)
}
