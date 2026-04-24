package server_test

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

// Compile-time assertions.
var (
	_ backend.Client             = (*server.Client)(nil)
	_ backend.ServerCapabilities = (*server.Client)(nil)
)
