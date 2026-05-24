package cloud_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// TestCloud_PAT_NotImplemented asserts that the Cloud client does NOT implement
// PATClient — AsPATClient must return ErrUnsupportedOnHost for Cloud backends.
func TestCloud_PAT_NotImplemented(t *testing.T) {
	t.Parallel()
	var c backend.Client = (*cloud.Client)(nil)
	_, ok := c.(backend.PATClient)
	if ok {
		t.Fatal("cloud.Client must not implement PATClient; Cloud PAT management is web-only")
	}

	_, err := backend.AsPATClient(c, "bitbucket.org")
	if err == nil {
		t.Fatal("AsPATClient should return an error for Cloud")
	}
	de, ok := err.(*backend.DomainError)
	if !ok {
		t.Fatalf("expected *backend.DomainError, got %T", err)
	}
	if de.Kind != backend.ErrUnsupportedOnHost {
		t.Fatalf("expected ErrUnsupportedOnHost, got %v", de.Kind)
	}
}
