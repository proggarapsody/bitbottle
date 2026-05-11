package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/errfmt"
)

type handlers struct {
	f *factory.Factory
}

func newHandlers(f *factory.Factory) *handlers {
	return &handlers{f: f}
}

// resolveBackend picks a host and dials a backend client. Host
// inference (single-host fallback, ambiguity errors) is delegated
// to factory.ResolveHost so the rule lives in exactly one place —
// the same place ResolveTarget consults for bare PROJECT/REPO args.
func (h *handlers) resolveBackend(hostname string) (backend.Client, error) {
	host, err := factory.ResolveHost(h.f, hostname)
	if err != nil {
		return nil, err
	}
	return h.f.Backend(host)
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

// errorEnvelope is the structured shape MCP clients receive on tool
// failures. Fields:
//
//   - Code:     dotted backend.ErrorCode token (e.g. "auth.invalid_token",
//     "repo.not_found"). When the underlying DomainError has no Code,
//     a kind-based fallback string ("auth", "not_found", "conflict",
//     "permission", "unsupported_on_host", "transport") is emitted so
//     clients always get a non-empty signal.
//   - Host/Feature/Resource/ID: optional context fields stamped by the
//     adapter at the call site.
//   - Message:  the human-readable error text (already control-byte
//     sanitised by errfmt's renderer for CLI; raw here for JSON).
//   - Hints:    actionable next-step strings sourced from errfmt's
//     catalogue, with template placeholders expanded against the
//     DomainError fields. AI-agent integrations surface these to the
//     user without bundling their own copy of the catalogue.
type errorEnvelope struct {
	Code     string   `json:"code"`
	Host     string   `json:"host,omitempty"`
	Feature  string   `json:"feature,omitempty"`
	Resource string   `json:"resource,omitempty"`
	ID       string   `json:"id,omitempty"`
	Message  string   `json:"message"`
	Hints    []string `json:"hints,omitempty"`
}

func errResultErr(err error) *mcplib.CallToolResult {
	var de *backend.DomainError
	if errors.As(err, &de) {
		env := errorEnvelope{
			Code:     envelopeCode(de),
			Host:     de.Host,
			Feature:  de.Feature,
			Resource: de.Resource,
			ID:       de.ID,
			Message:  de.Error(),
			Hints:    errfmt.HintsFor(de),
		}
		if data, mErr := json.Marshal(env); mErr == nil {
			return mcplib.NewToolResultError(string(data))
		}
	}
	return mcplib.NewToolResultError(err.Error())
}

// envelopeCode prefers the dotted ErrorCode (the join key with errfmt's
// catalogue) and falls back to a kind-based label when Code is unset, so
// MCP clients always get a structured token. Once every code path stamps
// a Code, the kind fallback can be retired.
func envelopeCode(de *backend.DomainError) string {
	if de.Code != "" {
		return string(de.Code)
	}
	switch {
	case errors.Is(de.Kind, backend.ErrNotFound):
		return "not_found"
	case errors.Is(de.Kind, backend.ErrAuth):
		return "auth"
	case errors.Is(de.Kind, backend.ErrPermission):
		return "permission"
	case errors.Is(de.Kind, backend.ErrUnsupportedOnHost):
		return "unsupported_on_host"
	case errors.Is(de.Kind, backend.ErrConflict):
		return "conflict"
	case errors.Is(de.Kind, backend.ErrTransport):
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
	namespace := req.GetString("namespace", "")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	repos, err := client.ListRepos(namespace, limit)
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

func (h *handlers) renameRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	newName, err := requireString(req, "new_name")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	repo, err := client.RenameRepo(project, slug, newName)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(repo)
}

func (h *handlers) forkRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	into, err := requireString(req, "into")
	if err != nil {
		return errResultErr(err), nil
	}
	name := req.GetString("name", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	forker, err := backend.AsRepoForker(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	fork, err := forker.ForkRepo(project, slug, backend.ForkRepoInput{
		Workspace: into,
		Name:      name,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(fork)
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

func (h *handlers) reopenPR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	reopener, err := backend.AsPRReopener(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := reopener.ReopenPR(project, slug, id); err != nil {
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
	if req.GetBool("inline_only", false) {
		filtered := make([]backend.PRComment, 0, len(cmts))
		for _, c := range cmts {
			if c.Inline != nil {
				filtered = append(filtered, c)
			}
		}
		cmts = filtered
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
	in := backend.AddPRCommentInput{Text: body}
	if inlinePath := req.GetString("inline_path", ""); inlinePath != "" {
		line := req.GetInt("inline_line", 0)
		if line <= 0 {
			return errResult("inline_path requires inline_line (positive integer)"), nil
		}
		side := req.GetString("inline_side", "new")
		if side != "new" && side != "old" {
			return errResult(`inline_side must be "new" or "old"`), nil
		}
		inline := &backend.PRCommentInline{Path: inlinePath, Side: side, Line: line}
		if startLine := req.GetInt("inline_start_line", 0); startLine != 0 {
			if startLine > line {
				return errResult("inline_start_line must be <= inline_line"), nil
			}
			inline.StartLine = startLine
		}
		in.Inline = inline
	}
	if parent := req.GetInt("parent_id", 0); parent > 0 {
		p := parent
		in.Parent = &p
	}
	c, err := client.AddPRComment(project, slug, id, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(c)
}

func (h *handlers) editPRComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResult("missing required parameter: comment_id"), nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	c, err := client.EditPRComment(project, slug, id, commentID, body)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(c)
}

func (h *handlers) deletePRComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResult("missing required parameter: comment_id"), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.DeletePRComment(project, slug, id, commentID); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

// submitPRReview implements the submit_pr_review tool — the MCP-side mirror
// of `pr review`. It takes an action ("approve" / "request_changes" /
// "comment"; default "comment" when body or inline_comments is set), an
// optional top-level body, and an optional inline_comments array of
// {path, line, body, [start_line], [side]} objects. Validation matches the
// CLI: at least one of action/body/inline_comments must be set, and each
// inline entry must carry a non-empty path, a positive line, and a body.
func (h *handlers) submitPRReview(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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

	action := req.GetString("action", "")
	body := req.GetString("body", "")

	inlineRaw, _ := req.GetArguments()["inline_comments"].([]any)
	inline := make([]backend.SubmitReviewInline, 0, len(inlineRaw))
	for i, item := range inlineRaw {
		obj, ok := item.(map[string]any)
		if !ok {
			return errResult(fmt.Sprintf("inline_comments[%d]: must be an object", i)), nil
		}
		ic, perr := parseInlineCommentObject(obj)
		if perr != nil {
			return errResult(fmt.Sprintf("inline_comments[%d]: %v", i, perr)), nil
		}
		inline = append(inline, ic)
	}

	if action == "" {
		if body == "" && len(inline) == 0 {
			return errResult(`one of action ("approve", "request_changes", "comment"), body, or inline_comments is required`), nil
		}
		action = "comment"
	}
	switch action {
	case "approve", "request_changes", "comment":
	default:
		return errResult(fmt.Sprintf(`action must be "approve", "request_changes", or "comment" (got %q)`, action)), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.SubmitReview(project, slug, id, backend.SubmitReviewInput{
		Action: action,
		Body:   body,
		Inline: inline,
	}); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
}

// parseInlineCommentObject pulls a SubmitReviewInline out of a raw JSON
// object, applying the same shape rules the CLI's parseInlineReviewSpec
// enforces: path/line/body required, side defaults to "new", start_line
// must be <= line when present.
func parseInlineCommentObject(obj map[string]any) (backend.SubmitReviewInline, error) {
	path, _ := obj["path"].(string)
	if path == "" {
		return backend.SubmitReviewInline{}, fmt.Errorf("path is required")
	}
	body, _ := obj["body"].(string)
	if body == "" {
		return backend.SubmitReviewInline{}, fmt.Errorf("body is required")
	}
	lineF, ok := obj["line"].(float64)
	if !ok || lineF <= 0 {
		return backend.SubmitReviewInline{}, fmt.Errorf("line must be a positive integer")
	}
	out := backend.SubmitReviewInline{
		Path: path,
		Line: int(lineF),
		Body: body,
		Side: "new",
	}
	if side, ok := obj["side"].(string); ok && side != "" {
		if side != "new" && side != "old" {
			return backend.SubmitReviewInline{}, fmt.Errorf(`side must be "new" or "old" (got %q)`, side)
		}
		out.Side = side
	}
	if startF, ok := obj["start_line"].(float64); ok && startF > 0 {
		start := int(startF)
		if start > out.Line {
			return backend.SubmitReviewInline{}, fmt.Errorf("start_line %d must be <= line %d", start, out.Line)
		}
		out.StartLine = start
	}
	return out, nil
}

func (h *handlers) resolvePRComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResult("missing required parameter: comment_id"), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	resolver, err := backend.AsPRCommentResolver(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := resolver.ResolvePRComment(project, slug, id, commentID); err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText("{}"), nil
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

func (h *handlers) listPipelineSteps(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	steps, err := pc.ListPipelineSteps(project, slug, uuid)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(steps)
}

func (h *handlers) getPipelineStepLog(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	pipelineUUID, err := requireString(req, "pipeline_uuid")
	if err != nil {
		return errResultErr(err), nil
	}
	stepUUID, err := requireString(req, "step_uuid")
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
	rc, err := pc.GetPipelineStepLog(project, slug, pipelineUUID, stepUUID)
	if err != nil {
		return errResultErr(err), nil
	}
	defer rc.Close() //nolint:errcheck
	body, err := io.ReadAll(rc)
	if err != nil {
		return errResultErr(err), nil
	}
	return mcplib.NewToolResultText(string(body)), nil
}

func (h *handlers) listPipelineVariables(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	pc, err := backend.AsPipelineClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	vars, err := pc.ListPipelineVariables(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	// Defensive: even though the wire layer never returns Value for secured
	// vars, blank it again before serialising in case a future API change
	// leaks one through.
	for i := range vars {
		if vars[i].Secured {
			vars[i].Value = ""
		}
	}
	return jsonResult(vars)
}

func (h *handlers) setPipelineVariable(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
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
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	pc, err := backend.AsPipelineClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	v, err := pc.SetPipelineVariable(project, slug, backend.PipelineVariableInput{
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

func (h *handlers) deletePipelineVariable(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	key, err := requireString(req, "key")
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
	if err := pc.DeletePipelineVariable(project, slug, key); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"key": key, "status": "deleted"})
}

func (h *handlers) listWebhooks(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
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
	hooks, err := client.ListWebhooks(project, slug)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hooks)
}

func (h *handlers) getWebhook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, err := requireString(req, "id")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	hook, err := client.GetWebhook(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hook)
}

func (h *handlers) createWebhook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	url, err := requireString(req, "url")
	if err != nil {
		return errResultErr(err), nil
	}
	events := req.GetStringSlice("events", nil)
	if len(events) == 0 {
		return errResult("events: required and must contain at least one event key"), nil
	}
	active := req.GetBool("active", true)
	secret := req.GetString("secret", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	hook, err := client.CreateWebhook(project, slug, backend.CreateWebhookInput{
		URL:    url,
		Events: events,
		Active: active,
		Secret: secret,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hook)
}

func (h *handlers) deleteWebhook(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, err := requireString(req, "id")
	if err != nil {
		return errResultErr(err), nil
	}
	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.DeleteWebhook(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]string{"id": id, "status": "deleted"})
}

func (h *handlers) listWorkspaces(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	wc, err := backend.AsWorkspaceClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	ws, err := wc.ListWorkspaces(limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(ws)
}

func (h *handlers) listProjects(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	limit := req.GetInt("limit", 30)

	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	wc, err := backend.AsWorkspaceClient(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	projects, err := wc.ListProjects(workspace, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(projects)
}

// resolveIssueClient is the small adapter every issue handler shares: pick
// the host, dial the backend, type-assert IssueClient, and gather the
// project/slug args. Keeps each handler body trivial.
func (h *handlers) resolveIssueClient(req mcplib.CallToolRequest) (backend.IssueClient, string, string, error) {
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
	ic, err := backend.AsIssueClient(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return ic, project, slug, nil
}

func (h *handlers) listIssues(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	state := req.GetString("state", "")
	limit := req.GetInt("limit", 30)
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	issues, err := ic.ListIssues(project, slug, state, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issues)
}

func (h *handlers) getIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	issue, err := ic.GetIssue(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issue)
}

func (h *handlers) createIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	title, err := requireString(req, "title")
	if err != nil {
		return errResultErr(err), nil
	}
	body := req.GetString("body", "")
	kind := req.GetString("kind", "")
	priority := req.GetString("priority", "")
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	issue, err := ic.CreateIssue(project, slug, backend.CreateIssueInput{
		Title:    title,
		Content:  body,
		Kind:     kind,
		Priority: priority,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issue)
}

func (h *handlers) closeIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	issue, err := ic.UpdateIssue(project, slug, id, backend.UpdateIssueInput{State: "closed"})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issue)
}

func (h *handlers) updateIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	in := backend.UpdateIssueInput{
		Title:    req.GetString("title", ""),
		Content:  req.GetString("body", ""),
		Kind:     req.GetString("kind", ""),
		Priority: req.GetString("priority", ""),
		Assignee: req.GetString("assignee", ""),
		State:    req.GetString("state", ""),
	}
	issue, err := ic.UpdateIssue(project, slug, id, in)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(issue)
}

func (h *handlers) reopenIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ic.ReopenIssue(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "state": "open"})
}

func (h *handlers) assignIssue(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	assignee, err := requireString(req, "assignee")
	if err != nil {
		return errResultErr(err), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ic.AssignIssue(project, slug, id, assignee); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "assignee": assignee})
}

func (h *handlers) listIssueComments(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	comments, err := ic.ListIssueComments(project, slug, id)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(comments)
}

func (h *handlers) addIssueComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	comment, err := ic.AddIssueComment(project, slug, id, body)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(comment)
}

func (h *handlers) editIssueComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: comment_id")), nil
	}
	body, err := requireString(req, "body")
	if err != nil {
		return errResultErr(err), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	comment, err := ic.EditIssueComment(project, slug, id, commentID, body)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(comment)
}

func (h *handlers) deleteIssueComment(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	commentID := req.GetInt("comment_id", 0)
	if commentID == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: comment_id")), nil
	}
	ic, project, slug, err := h.resolveIssueClient(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := ic.DeleteIssueComment(project, slug, id, commentID); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"deleted": true, "comment_id": commentID})
}

// resolveBranchProtector mirrors resolveIssueClient: pick the host, dial the
// backend, type-assert BranchProtector, and gather the project/slug args.
func (h *handlers) resolveBranchProtector(req mcplib.CallToolRequest) (backend.BranchProtector, string, string, error) {
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
	bp, err := backend.AsBranchProtector(client, hostname)
	if err != nil {
		return nil, "", "", err
	}
	return bp, project, slug, nil
}

func (h *handlers) listBranchProtections(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	limit := req.GetInt("limit", 30)
	bp, project, slug, err := h.resolveBranchProtector(req)
	if err != nil {
		return errResultErr(err), nil
	}
	got, err := bp.ListBranchProtections(project, slug, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(got)
}

func (h *handlers) createBranchProtection(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	typ, err := requireString(req, "type")
	if err != nil {
		return errResultErr(err), nil
	}
	branch := req.GetString("branch", "")
	pattern := req.GetString("pattern", "")
	if (branch == "") == (pattern == "") {
		return errResult("specify exactly one of branch or pattern"), nil
	}
	matcherID := branch
	matcherKind := "BRANCH"
	if pattern != "" {
		matcherID = pattern
		matcherKind = "PATTERN"
	}
	users := req.GetStringSlice("users", nil)
	groups := req.GetStringSlice("groups", nil)

	bp, project, slug, err := h.resolveBranchProtector(req)
	if err != nil {
		return errResultErr(err), nil
	}
	out, err := bp.CreateBranchProtection(project, slug, backend.CreateBranchProtectionInput{
		Type:        typ,
		MatcherID:   matcherID,
		MatcherKind: matcherKind,
		Users:       users,
		Groups:      groups,
	})
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(out)
}

func (h *handlers) deleteBranchProtection(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	id := req.GetInt("id", 0)
	if id == 0 {
		return errResultErr(fmt.Errorf("missing required parameter: id")), nil
	}
	bp, project, slug, err := h.resolveBranchProtector(req)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := bp.DeleteBranchProtection(project, slug, id); err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(map[string]any{"id": id, "status": "deleted"})
}

// searchCode is the MCP handler for the `search_code` tool. Cloud-only —
// the AsCodeSearcher type-assertion stamps the host.unsupported envelope
// when the backend is Server / DC.
func (h *handlers) searchCode(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	workspace, err := requireString(req, "workspace")
	if err != nil {
		return errResultErr(err), nil
	}
	query, err := requireString(req, "query")
	if err != nil {
		return errResultErr(err), nil
	}
	limit := req.GetInt("limit", 30)

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	cs, err := backend.AsCodeSearcher(client, hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	hits, err := cs.SearchCode(workspace, query, limit)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(hits)
}
