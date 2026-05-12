package backend_test

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
	"github.com/proggarapsody/bitbottle/api/server"
)

// Compile-time assertions: server.Client and cloud.Client satisfy backend.Client.
var (
	_ backend.Client = (*server.Client)(nil)
	_ backend.Client = (*cloud.Client)(nil)
)
