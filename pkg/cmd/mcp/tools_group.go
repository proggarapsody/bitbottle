package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerGroupTools)
}

func registerGroupTools(s *mcpserver.MCPServer, h *handlers) {
	s.AddTool(
		mcplib.NewTool("list_groups",
			mcplib.WithDescription("List Bitbucket Server/DC admin groups. Returns an error on Cloud."),
			mcplib.WithString("filter",
				mcplib.Description("Filter groups by name prefix"),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of groups to return (default 100)"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.listGroups,
	)
	s.AddTool(
		mcplib.NewTool("create_group",
			mcplib.WithDescription("Create a Bitbucket Server/DC admin group. Returns an error on Cloud."),
			mcplib.WithString("name",
				mcplib.Required(),
				mcplib.Description("Name of the group to create"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.createGroup,
	)
	s.AddTool(
		mcplib.NewTool("delete_group",
			mcplib.WithDescription("Delete a Bitbucket Server/DC admin group. Returns an error on Cloud."),
			mcplib.WithString("name",
				mcplib.Required(),
				mcplib.Description("Name of the group to delete"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.deleteGroup,
	)
	s.AddTool(
		mcplib.NewTool("list_group_members",
			mcplib.WithDescription("List members of a Bitbucket Server/DC admin group. Returns an error on Cloud."),
			mcplib.WithString("group",
				mcplib.Required(),
				mcplib.Description("Name of the group"),
			),
			mcplib.WithNumber("limit",
				mcplib.Description("Maximum number of members to return (default 100)"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.listGroupMembers,
	)
	s.AddTool(
		mcplib.NewTool("add_group_member",
			mcplib.WithDescription("Add a user to a Bitbucket Server/DC admin group. Returns an error on Cloud."),
			mcplib.WithString("group",
				mcplib.Required(),
				mcplib.Description("Name of the group"),
			),
			mcplib.WithString("user",
				mcplib.Required(),
				mcplib.Description("Username (slug) to add"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.addGroupMember,
	)
	s.AddTool(
		mcplib.NewTool("remove_group_member",
			mcplib.WithDescription("Remove a user from a Bitbucket Server/DC admin group. Returns an error on Cloud."),
			mcplib.WithString("group",
				mcplib.Required(),
				mcplib.Description("Name of the group"),
			),
			mcplib.WithString("user",
				mcplib.Required(),
				mcplib.Description("Username (slug) to remove"),
			),
			mcplib.WithString("hostname",
				mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
			),
		),
		h.removeGroupMember,
	)
}
