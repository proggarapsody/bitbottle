package mcp

import (
	"context"
	"encoding/json"
	"errors"
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

type errorEnvelope struct {
	Code     string `json:"code"`
	Host     string `json:"host,omitempty"`
	Feature  string `json:"feature,omitempty"`
	Resource string `json:"resource,omitempty"`
	ID       string `json:"id,omitempty"`
	Message  string `json:"message"`
}

func errResultErr(err error) *mcplib.CallToolResult {
	var de *backend.DomainError
	if errors.As(err, &de) {
		env := errorEnvelope{
			Code:     domainErrorCode(de.Kind),
			Host:     de.Host,
			Feature:  de.Feature,
			Resource: de.Resource,
			ID:       de.ID,
			Message:  de.Error(),
		}
		if data, mErr := json.Marshal(env); mErr == nil {
			return mcplib.NewToolResultError(string(data))
		}
	}
	return mcplib.NewToolResultError(err.Error())
}

func domainErrorCode(kind error) string {
	switch {
	case errors.Is(kind, backend.ErrNotFound):
		return "not_found"
	case errors.Is(kind, backend.ErrAuth):
		return "auth"
	case errors.Is(kind, backend.ErrPermission):
		return "permission"
	case errors.Is(kind, backend.ErrUnsupportedOnHost):
		return "unsupported_on_host"
	case errors.Is(kind, backend.ErrConflict):
		return "conflict"
	case errors.Is(kind, backend.ErrTransport):
		return "transport"
	default:
		return "error"
	}
}

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
		return errResultErr(err), nil
	}
	return jsonResult(cfg.Hosts())
}

func (h *handlers) listRepos(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	repos, err := client.ListRepos(limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repos)
}

func (h *handlers) getRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := client.GetRepo(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repo)
}

func (h *handlers) createRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	description := req.GetString("description", "")
	private := req.GetBool("private", false)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := client.CreateRepo(project, backend.CreateRepoInput{
		Name:        name,
		SCM:         "git",
		Public:      !private,
		Description: description,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repo)
}

func (h *handlers) deleteRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.DeleteRepo(project, slug); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) listPRs(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	state := req.GetString("state", "OPEN")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	prs, err := client.ListPRs(project, slug, state, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(prs)
}

func (h *handlers) getPR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pr, err := client.GetPR(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pr)
}

func (h *handlers) createPR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	title, err := requireString(req, "title")
	if err != nil {
		return errResultErr(err), nil
	}
	fromBranch, err := requireString(req, "from_branch")
	if err != nil {
		return errResultErr(err), nil
	}
	toBranch, err := requireString(req, "to_branch")
	if err != nil {
		return errResultErr(err), nil
	}
	body := req.GetString("body", "")
	draft := req.GetBool("draft", false)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pr, err := client.CreatePR(project, slug, backend.CreatePRInput{
		Title:       title,
		Description: body,
		Draft:       draft,
		FromBranch:  fromBranch,
		ToBranch:    toBranch,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pr)
}

func (h *handlers) mergePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}
	strategy := req.GetString("strategy", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pr, err := client.MergePR(project, slug, id, backend.MergePRInput{Strategy: strategy})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pr)
}

func (h *handlers) approvePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.ApprovePR(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) getPRDiff(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	diff, err := client.GetPRDiff(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(diff), nil
}

func (h *handlers) deleteBranch(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
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
	if err := client.DeleteBranch(project, slug, branch); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) getCurrentUser(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	user, err := client.GetCurrentUser()
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(user)
}

func (h *handlers) listBranches(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	branches, err := client.ListBranches(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(branches)
}

func (h *handlers) listPipelines(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 20)

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsPipelineClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pipelines, err := pc.ListPipelines(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pipelines)
}

func (h *handlers) getPipeline(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	uuid, err := requireString(req, "uuid")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsPipelineClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pl, err := pc.GetPipeline(project, slug, uuid)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pl)
}

func (h *handlers) createBranch(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	startAt, err := requireString(req, "start_at")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	br, err := client.CreateBranch(project, slug, backend.CreateBranchInput{
		Name:    name,
		StartAt: startAt,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(br)
}

func (h *handlers) listTags(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	tags, err := client.ListTags(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(tags)
}

func (h *handlers) createTag(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}
	startAt, err := requireString(req, "start_at")
	if err != nil {
		return errResultErr(err), nil
	}
	message := req.GetString("message", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	t, err := client.CreateTag(project, slug, backend.CreateTagInput{
		Name:    name,
		StartAt: startAt,
		Message: message,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(t)
}

func (h *handlers) deleteTag(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	name, err := requireString(req, "name")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.DeleteTag(project, slug, name); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) updatePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}
	title := req.GetString("title", "")
	body := req.GetString("body", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pr, err := client.UpdatePR(project, slug, id, backend.UpdatePRInput{
		Title:       title,
		Description: body,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pr)
}

func (h *handlers) declinePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.DeclinePR(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) unapprovePR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.UnapprovePR(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) readyPR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.ReadyPR(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	pr, err := client.GetPR(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pr)
}

func (h *handlers) requestReview(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}
	reviewers, err := requireString(req, "reviewers")
	if err != nil {
		return errResultErr(err), nil
	}

	var users []string
	for _, u := range splitTrimmed(reviewers, ",") {
		if u != "" {
			users = append(users, u)
		}
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.RequestReview(project, slug, id, users); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

func (h *handlers) listCommits(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	branch := req.GetString("branch", "main")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	commits, err := client.ListCommits(project, slug, branch, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(commits)
}

func (h *handlers) getCommit(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	commit, err := client.GetCommit(project, slug, hash)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(commit)
}

func (h *handlers) runPipeline(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")

	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
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
	pc, err := backend.AsPipelineClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pl, err := pc.RunPipeline(project, slug, backend.RunPipelineInput{Branch: branch})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(pl)
}

func (h *handlers) listPRComments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	cmts, err := client.ListPRComments(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(cmts)
}

func (h *handlers) addPRComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResult("missing required parameter: id"), nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	c, err := client.AddPRComment(project, slug, id, backend.AddPRCommentInput{Text: body})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(c)
}

func (h *handlers) listCommitStatuses(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	statuses, err := client.ListCommitStatuses(project, slug, hash)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(statuses)
}
