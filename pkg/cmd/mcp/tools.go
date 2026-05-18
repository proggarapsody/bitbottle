package mcp

import (
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func newMCPServer(f *factory.Factory) *mcpserver.MCPServer {
	s := mcpserver.NewMCPServer("bitbottle", "0.1.0",
		mcpserver.WithToolCapabilities(false),
	)
	h := newHandlers(f)
	registerTools(s, h)
	return s
}

func registerTools(s *mcpserver.MCPServer, h *handlers) {
	// All tools are registered via init() functions in per-domain files
	// using registerTool(). This function just fires all of them.
	for _, fn := range registeredFns() {
		fn(s, h)
	}
}
