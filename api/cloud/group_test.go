package cloud_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// TestCloud_GroupClient_NotImplemented verifies Cloud does not implement GroupClient.
func TestCloud_GroupClient_NotImplemented(t *testing.T) {
	t.Parallel()
	var c any = &cloud.Client{}
	if _, ok := c.(backend.GroupClient); ok {
		t.Error("Cloud client must not implement GroupClient; use AsGroupClient which returns ErrUnsupportedOnHost")
	}
}

// TestCloud_GroupMemberClient_NotImplemented verifies Cloud does not implement GroupMemberClient.
func TestCloud_GroupMemberClient_NotImplemented(t *testing.T) {
	t.Parallel()
	var c any = &cloud.Client{}
	if _, ok := c.(backend.GroupMemberClient); ok {
		t.Error("Cloud client must not implement GroupMemberClient; use AsGroupMemberClient which returns ErrUnsupportedOnHost")
	}
}
