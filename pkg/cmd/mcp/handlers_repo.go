package mcp

import (
	"context"
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/git"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func (h *handlers) listRepos(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	namespace := req.GetString("namespace", "")
	limit := req.GetInt("limit", 30)
	if err := validateRange("limit", limit, 1, 100); err != nil {
		return errResult(err.Error()), nil
	}

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

func (h *handlers) cloneRepo(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	dir := req.GetString("dir", slug)
	protocol := req.GetString("protocol", "ssh")

	host, err := factory.ResolveHost(h.f, hostname)
	if err != nil {
		return errResultErr(err), nil
	}

	// Resolve clone URL via API; fall back to heuristic.
	cloneURL := resolveCloneURLMCP(h.f, host, project, slug, protocol)

	g := git.New(h.f.GitRunner())
	if err := g.Clone(cloneURL, dir); err != nil {
		return errResultErr(err), nil
	}

	_ = g.SetConfigInDir(dir, "bitbottle.host", host)
	_ = g.SetConfigInDir(dir, "bitbottle.project", project)
	_ = g.SetConfigInDir(dir, "bitbottle.slug", slug)

	return mcplib.NewToolResultText(fmt.Sprintf("cloned to %s", dir)), nil
}

// resolveCloneURLMCP picks the clone URL from the API or falls back to heuristic.
func resolveCloneURLMCP(f *factory.Factory, host, project, slug, protocol string) string {
	client, err := f.Backend(host)
	if err == nil {
		repo, err := client.GetRepo(project, slug)
		if err == nil && len(repo.CloneURLs) > 0 {
			useSSH := protocol == "ssh" || protocol == ""
			if useSSH {
				for _, u := range repo.CloneURLs {
					if u.Name == "ssh" {
						return u.URL
					}
				}
			}
			for _, u := range repo.CloneURLs {
				if u.Name == "https" || u.Name == "http" {
					return u.URL
				}
			}
			return repo.CloneURLs[0].URL
		}
	}
	// Heuristic fallback.
	isCloud := false
	cfg, _ := f.Config()
	if cfg != nil {
		hc, _ := cfg.Get(host)
		isCloud = bbinstance.IsCloud(host, hc.BackendType)
	}
	if protocol == "ssh" || protocol == "" {
		if isCloud {
			return bbinstance.CloudSSHURL(project, slug)
		}
		return fmt.Sprintf("ssh://git@%s/%s/%s.git", host, project, slug)
	}
	if isCloud {
		return bbinstance.CloudHTTPSURL(project, slug)
	}
	return bbinstance.HTTPSURL(host, project, slug)
}
