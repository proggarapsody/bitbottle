package backend_test

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
	"github.com/proggarapsody/bitbottle/api/server"
)

// Compile-time assertions: server.Client and cloud.Client satisfy backend.Client.
// server.Client also satisfies backend.ServerCapabilities.
// cloud.Client does NOT satisfy backend.ServerCapabilities (no GetApplicationProperties).
var (
	_ backend.Client             = (*server.Client)(nil)
	_ backend.Client             = (*cloud.Client)(nil)
	_ backend.ServerCapabilities = (*server.Client)(nil)
)
