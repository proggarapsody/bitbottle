package cloud_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// TestCloud_ServerProjectClient_NotImplemented verifies Cloud does not implement ServerProjectClient.
func TestCloud_ServerProjectClient_NotImplemented(t *testing.T) {
	t.Parallel()
	var c any = &cloud.Client{}
	if _, ok := c.(backend.ServerProjectClient); ok {
		t.Error("Cloud client must not implement ServerProjectClient; use AsServerProjectClient which returns ErrUnsupportedOnHost")
	}
}
