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

// Compile-time assertion: cloud.Client satisfies backend.DeploymentClient.
var _ backend.DeploymentClient = (*cloud.Client)(nil)

// Compile-time assertions: both cloud.Client and server.Client satisfy backend.DiffClient.
var (
	_ backend.DiffClient = (*cloud.Client)(nil)
	_ backend.DiffClient = (*server.Client)(nil)
)
