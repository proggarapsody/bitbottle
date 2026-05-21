package server_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestServerClient_StopPipeline_Unsupported(t *testing.T) {
	t.Parallel()
	c, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		// StopPipeline must not reach the network — no HTTP call expected.
		t.Error("unexpected HTTP request")
	})
	err := c.StopPipeline("proj", "repo", "uuid")
	var de *backend.DomainError
	if !errors.As(err, &de) || !errors.Is(de.Kind, backend.ErrUnsupportedOnHost) {
		t.Fatalf("expected ErrUnsupportedOnHost, got %v", err)
	}
}
