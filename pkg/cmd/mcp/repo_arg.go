package mcp

import (
	"fmt"

	mcplib "github.com/mark3labs/mcp-go/mcp"
)

// deprecationNote is the prefix prepended (as a separate text block) to a
// tool result when the caller used a deprecated argument shape. MCP-04 keeps
// one release of back-compat: the old shape still works, but the result text
// nudges clients toward the canonical {project, slug} shape.
const deprecationNote = "DEPRECATION: this tool now uses {project, slug}; the old argument shape is accepted for one release and will be removed. Migrate to {project, slug}."

// repoFromProjectSlugOrRepo resolves the canonical {project, slug} repo-arg
// shape, falling back to the legacy single `repo` ("WORKSPACE/REPO") arg.
//
// Returns (ns, slug, deprecated, error). deprecated is true when the legacy
// `repo` arg was used so the handler can stamp a deprecation note. When both
// new and old shapes are absent, a "missing project/slug" error is returned.
func repoFromProjectSlugOrRepo(req mcplib.CallToolRequest) (ns, slug string, deprecated bool, err error) {
	project := req.GetString("project", "")
	s := req.GetString("slug", "")
	if project != "" && s != "" {
		return project, s, false, nil
	}
	if legacy := req.GetString("repo", ""); legacy != "" {
		n, sl, e := splitRepo(legacy)
		if e != nil {
			return "", "", true, e
		}
		return n, sl, true, nil
	}
	return "", "", false, fmt.Errorf("missing required parameters: project and slug")
}

// repoFromProjectSlugOrProjectRepo resolves the canonical {project, slug}
// repo-arg shape, falling back to the legacy {project, repo} shape (where
// `repo` is a bare slug, not WORKSPACE/REPO).
//
// Returns (project, slug, deprecated, error). deprecated is true when the
// legacy `repo` arg supplied the slug.
func repoFromProjectSlugOrProjectRepo(req mcplib.CallToolRequest) (project, slug string, deprecated bool, err error) {
	project = req.GetString("project", "")
	if project == "" {
		return "", "", false, fmt.Errorf("missing required parameter: project")
	}
	if s := req.GetString("slug", ""); s != "" {
		return project, s, false, nil
	}
	if legacy := req.GetString("repo", ""); legacy != "" {
		return project, legacy, true, nil
	}
	return "", "", false, fmt.Errorf("missing required parameter: slug")
}

// withDeprecation wraps a successful tool result, prepending a deprecation
// note text block when deprecated is true. On error results it is a no-op so
// the structured error envelope reaches the client unchanged.
func withDeprecation(res *mcplib.CallToolResult, deprecated bool) *mcplib.CallToolResult {
	if !deprecated || res == nil || res.IsError {
		return res
	}
	notice := mcplib.NewToolResultText(deprecationNote)
	res.Content = append(notice.Content, res.Content...)
	return res
}
