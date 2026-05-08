package mcp

import (
	"context"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func init() {
	registerTool(registerSourceTools)
}

// registerSourceTools wires the RV1 source primitives — get_file_content
// and list_tree — onto the MCP server. Both backends implement
// SourceReader so neither tool needs an As* type-assertion.
func registerSourceTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("get_file_content",
			mcplib.WithDescription("Read a file's bytes at a given ref (branch, tag, or commit hash) without cloning. Output is base64-decoded by the client where binary; UTF-8 strings round-trip cleanly."),
			mcplib.WithString("project", mcplib.Required(), mcplib.Description("Project key (Server) or workspace slug (Cloud)")),
			mcplib.WithString("slug", mcplib.Required(), mcplib.Description("Repository slug")),
			mcplib.WithString("ref", mcplib.Required(), mcplib.Description("Branch, tag, or commit hash")),
			mcplib.WithString("path", mcplib.Required(), mcplib.Description("Repo-relative file path")),
			mcplib.WithString("hostname", mcplib.Description("Bitbucket hostname (omit when only one host is configured)")),
		),
		h.getFileContent,
	)
	s.AddTool(
		mcplib.NewTool("list_tree",
			mcplib.WithDescription("List the immediate children of a directory at a ref. Each entry has type 'file' or 'dir' and a full repo-relative path. Submodules surface as 'dir' so callers can recurse uniformly."),
			mcplib.WithString("project", mcplib.Required(), mcplib.Description("Project key (Server) or workspace slug (Cloud)")),
			mcplib.WithString("slug", mcplib.Required(), mcplib.Description("Repository slug")),
			mcplib.WithString("ref", mcplib.Required(), mcplib.Description("Branch, tag, or commit hash")),
			mcplib.WithString("path", mcplib.Description("Repo-relative directory path; omit for the repository root")),
			mcplib.WithString("hostname", mcplib.Description("Bitbucket hostname (omit when only one host is configured)")),
		),
		h.listTree,
	)
}

func (h *handlers) getFileContent(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	ref, err := requireString(req, "ref")
	if err != nil {
		return errResultErr(err), nil
	}
	path, err := requireString(req, "path")
	if err != nil {
		return errResultErr(err), nil
	}

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	body, err := client.GetFileContent(project, slug, ref, path)
	if err != nil {
		return errResultErr(err), nil
	}
	// Bytes returned as text; binary callers can re-encode if needed. Keeping
	// the envelope text-only (rather than base64) matches gh's `gh api` shape
	// and is simpler for the common case (source code).
	return mcplib.NewToolResultText(string(body)), nil
}

func (h *handlers) listTree(_ context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
	hostname := req.GetString("hostname", "")
	project, err := requireString(req, "project")
	if err != nil {
		return errResultErr(err), nil
	}
	slug, err := requireString(req, "slug")
	if err != nil {
		return errResultErr(err), nil
	}
	ref, err := requireString(req, "ref")
	if err != nil {
		return errResultErr(err), nil
	}
	// path is optional — empty means root.
	path := req.GetString("path", "")

	client, err := h.resolveBackend(hostname)
	if err != nil {
		return errResultErr(err), nil
	}
	entries, err := client.ListTree(project, slug, ref, path)
	if err != nil {
		return errResultErr(err), nil
	}
	return jsonResult(entries)
}

// Compile-time guard that the SourceReader interface stays reachable from
// this package — drift in the import set is the most common breakage when
// the interface evolves.
var _ backend.SourceReader = (backend.SourceReader)(nil)
