package cloud_test

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// Compile-time assertion: cloud.Client satisfies backend.Client.
var _ backend.Client = (*cloud.Client)(nil)

// Compile-time assertion: cloud.Client satisfies backend.PRChangesRequester.
var _ backend.PRChangesRequester = (*cloud.Client)(nil)
