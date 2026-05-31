package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/mcp/argval"
)

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
	if err := validateEnum("state", state, "OPEN", "MERGED", "DECLINED"); err != nil {
		return errResult(err.Error()), nil
	}
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	// MCP-09: strategy is optional; empty means "use the server default".
	// Only non-empty values are validated, and the allowed set never
	// contains "" (so the error can't read "must be one of , merge, ...").
	strategy, sErr := argval.EnumOneOf(req.GetArguments(), "strategy", []string{"merge", "squash", "rebase"})
	if sErr != nil {
		return errResultArg(sErr), nil
	}
	auto := req.GetBool("auto", false)
	autoStrategy := req.GetString("auto_strategy", "merge")
	if asErr := validateEnum("auto_strategy", autoStrategy, "merge", "squash", "rebase"); asErr != nil {
		return errResult(asErr.Error()), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	if auto {
		if err := client.EnableAutoMerge(project, slug, id, autoStrategy); err != nil {
			return errResultErr(err), nil
		}
		return mcplib.NewToolResultText(fmt.Sprintf(`{"queued":true,"strategy":%q}`, autoStrategy)), nil
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}
	// MCP-13: a no-op update (neither title nor body) is rejected client-side
	// rather than round-tripping to the API.
	if noOpErr := argval.OneOfRequired(req.GetArguments(), "title", "body"); noOpErr != nil {
		return errResultArg(noOpErr), nil
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
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

func (h *handlers) unreadyPR(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	if err := client.UnreadyPR(project, slug, id); err != nil {
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
	id, idErr := requireIntArg(req, "id")
	if idErr != nil {
		return idErr, nil
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
