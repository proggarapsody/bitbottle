package mcp

import (
	"sync"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

// toolRegistrarFn is a function that registers one or more MCP tools onto s
// using h as the handler receiver. It mirrors the signature used inside
// registerTools so per-domain files can contribute tools without touching the
// central tools.go.
type toolRegistrarFn func(s *mcpserver.MCPServer, h *handlers)

var (
	mu         sync.Mutex
	toolRegFns []toolRegistrarFn
)

// registerTool appends a registrar to the package-level slice. Call from
// init() in any file under pkg/cmd/mcp/.
func registerTool(fn toolRegistrarFn) {
	mu.Lock()
	defer mu.Unlock()
	toolRegFns = append(toolRegFns, fn)
}

// registeredFns returns a snapshot of all registrar functions contributed via
// registerTool. Exposed for testing.
func registeredFns() []toolRegistrarFn {
	mu.Lock()
	defer mu.Unlock()
	out := make([]toolRegistrarFn, len(toolRegFns))
	copy(out, toolRegFns)
	return out
}
