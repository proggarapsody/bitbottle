package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoDownloadTools)
}

func registerRepoDownloadTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Workspace slug"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_repo_downloads",
			mcplib.WithDescription("List repository download artifacts (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of downloads (default 30)"),
			),
		),
		h.listRepoDownloads,
	)

	s.AddTool(
		mcplib.NewTool("upload_repo_download",
			mcplib.WithDescription("Upload a file as a repository download artifact (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Name for the download artifact"),
				mcplib.Required(),
			),
			mcplib.WithString("file_content_base64",
				mcplib.Description("Base64-encoded file content to upload"),
				mcplib.Required(),
			),
		),
		h.uploadRepoDownload,
	)

	s.AddTool(
		mcplib.NewTool("delete_repo_download",
			mcplib.WithDescription("Delete a repository download artifact (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("name",
				mcplib.Description("Name of the download artifact to delete"),
				mcplib.Required(),
			),
		),
		h.deleteRepoDownload,
	)
}
