package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerRepoHookTools)
}

// registerRepoHookTools wires the repo hook tools onto the MCP server.
// All tools are Bitbucket Server / DC only — calls against Cloud return
// host.unsupported via AsRepoHookClient.
func registerRepoHookTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Project key"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)
	reqHookKey := mcplib.WithString("hook_key",
		mcplib.Description("Plugin hook key (e.g. com.example.plugin:my-hook)"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("list_repo_hooks",
			mcplib.WithDescription("List plugin hook scripts installed on a repository (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug,
		),
		h.listRepoHooks,
	)

	s.AddTool(
		mcplib.NewTool("view_repo_hook",
			mcplib.WithDescription("Get details for a single plugin hook script (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHookKey,
		),
		h.viewRepoHook,
	)

	s.AddTool(
		mcplib.NewTool("enable_repo_hook",
			mcplib.WithDescription("Enable a plugin hook script on a repository (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHookKey,
		),
		h.enableRepoHook,
	)

	s.AddTool(
		mcplib.NewTool("disable_repo_hook",
			mcplib.WithDescription("Disable a plugin hook script on a repository (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHookKey,
		),
		h.disableRepoHook,
	)

	s.AddTool(
		mcplib.NewTool("get_repo_hook_settings",
			mcplib.WithDescription("Get raw JSON settings for a plugin hook script (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHookKey,
		),
		h.getRepoHookSettings,
	)

	s.AddTool(
		mcplib.NewTool("set_repo_hook_settings",
			mcplib.WithDescription("Set JSON settings for a plugin hook script (Bitbucket Server / DC only)"),
			optHostname, reqProject, reqSlug, reqHookKey,
			mcplib.WithString("config",
				mcplib.Description("JSON string containing the hook settings"),
				mcplib.Required(),
			),
		),
		h.setRepoHookSettings,
	)
}
