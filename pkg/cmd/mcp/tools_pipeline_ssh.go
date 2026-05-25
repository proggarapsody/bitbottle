package mcp

import (
	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func init() {
	registerTool(registerPipelineSSHTools)
}

func registerPipelineSSHTools(s *mcpserver.MCPServer, h *handlers) {
	optHostname := mcplib.WithString("hostname",
		mcplib.Description("Bitbucket hostname (omit when only one host is configured)"),
	)
	reqProject := mcplib.WithString("project",
		mcplib.Description("Workspace or project key"),
		mcplib.Required(),
	)
	reqSlug := mcplib.WithString("slug",
		mcplib.Description("Repository slug"),
		mcplib.Required(),
	)
	reqUUID := mcplib.WithString("uuid",
		mcplib.Description("Known host UUID"),
		mcplib.Required(),
	)

	s.AddTool(
		mcplib.NewTool("view_pipeline_ssh_key_pair",
			mcplib.WithDescription("View the pipeline SSH key pair for a repository (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.viewPipelineSSHKeyPair,
	)

	s.AddTool(
		mcplib.NewTool("regenerate_pipeline_ssh_key_pair",
			mcplib.WithDescription("Regenerate the pipeline SSH key pair for a repository (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithNumber("bits",
				mcplib.Description("Key size in bits (2048 or 4096; default 2048)"),
			),
		),
		h.regeneratePipelineSSHKeyPair,
	)

	s.AddTool(
		mcplib.NewTool("list_pipeline_known_hosts",
			mcplib.WithDescription("List pipeline known hosts for a repository (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
		),
		h.listPipelineKnownHosts,
	)

	s.AddTool(
		mcplib.NewTool("view_pipeline_known_host",
			mcplib.WithDescription("View a pipeline known host by UUID (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqUUID,
		),
		h.viewPipelineKnownHost,
	)

	s.AddTool(
		mcplib.NewTool("add_pipeline_known_host",
			mcplib.WithDescription("Add a known host to pipeline SSH config (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			mcplib.WithString("hostname_arg",
				mcplib.Description("Hostname to add (e.g. github.com)"),
				mcplib.Required(),
			),
			mcplib.WithString("key",
				mcplib.Description("Base64-encoded public key material"),
			),
			mcplib.WithString("key_type",
				mcplib.Description("Key type: RSA, ECDSA, or Ed25519 (default RSA)"),
			),
		),
		h.addPipelineKnownHost,
	)

	s.AddTool(
		mcplib.NewTool("delete_pipeline_known_host",
			mcplib.WithDescription("Delete a pipeline known host by UUID (Cloud only)"),
			optHostname,
			reqProject,
			reqSlug,
			reqUUID,
		),
		h.deletePipelineKnownHost,
	)
}
