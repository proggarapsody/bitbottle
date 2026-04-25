package mcp

import (
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdMCP(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server commands",
	}
	cmd.AddCommand(newCmdMCPServe(f))
	return cmd
}

func newCmdMCPServe(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP stdio server",
		Long:  "Start a Model Context Protocol server over stdio. Add to Claude Desktop or Claude Code MCP config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s := newMCPServer(f)
			return mcpserver.ServeStdio(s)
		},
	}
}
