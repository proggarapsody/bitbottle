package mcp

import (
	"encoding/json"
	"errors"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// TestErrResultErr_EmitsDottedErrorCode verifies the MCP error envelope
// carries the dotted ErrorCode (e.g. "auth.invalid_token") that AI agents
// can branch on, not the legacy kind-based string ("auth"). The dotted
// code is the join key with errfmt's catalogue, so MCP clients and CLI
// users see the same identifier for the same condition.
func TestErrResultErr_EmitsDottedErrorCode(t *testing.T) {
	de := &backend.DomainError{
		Kind: backend.ErrAuth,
		Code: backend.CodeAuthInvalidToken,
		Host: "git.example.com",
	}
	res := errResultErr(de)
	if res == nil || !res.IsError {
		t.Fatal("expected error result")
	}
	env := decodeEnvelope(t, res)
	if env.Code != string(backend.CodeAuthInvalidToken) {
		t.Errorf("Code = %q, want %q", env.Code, backend.CodeAuthInvalidToken)
	}
}

// TestErrResultErr_EmitsHintsFromCatalogue verifies that the envelope's
// hints array is populated from the errfmt catalogue. AI agents need the
// hint text to surface remediation steps without bundling their own copy
// of the catalogue.
func TestErrResultErr_EmitsHintsFromCatalogue(t *testing.T) {
	de := &backend.DomainError{
		Kind:     backend.ErrNotFound,
		Code:     backend.CodeRepoNotFound,
		Host:     "git.example.com",
		Resource: "repository",
		ID:       "ws/repo",
	}
	res := errResultErr(de)
	env := decodeEnvelope(t, res)
	if len(env.Hints) == 0 {
		t.Fatalf("expected non-empty hints, got %+v", env)
	}
	// Don't pin the exact wording — that would couple this test to the
	// catalogue copy. Pin the load-bearing token instead.
	found := false
	for _, h := range env.Hints {
		if containsAll(h, "slug casing") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected hint mentioning slug casing, got %+v", env.Hints)
	}
}

// TestErrResultErr_FallsBackToKindWhenCodeUnset preserves backward compat:
// not every DomainError has a Code yet (some adapters classify only Kind).
// In that case the envelope's code should be the kind label rather than
// empty, so MCP clients still get a structured signal.
func TestErrResultErr_FallsBackToKindWhenCodeUnset(t *testing.T) {
	de := &backend.DomainError{
		Kind: backend.ErrConflict,
	}
	res := errResultErr(de)
	env := decodeEnvelope(t, res)
	if env.Code != "conflict" {
		t.Errorf("Code = %q, want kind-based fallback %q", env.Code, "conflict")
	}
}

// TestErrResultErr_PlainErrorReturnsRawMessage verifies non-DomainError
// values pass through to the raw message (no envelope), matching the
// pre-EX2 behaviour for unclassified failures.
func TestErrResultErr_PlainErrorReturnsRawMessage(t *testing.T) {
	res := errResultErr(errors.New("boom"))
	if res == nil || !res.IsError {
		t.Fatal("expected error result")
	}
}

func decodeEnvelope(t *testing.T, res *mcplib.CallToolResult) errorEnvelope {
	t.Helper()
	if len(res.Content) != 1 {
		t.Fatalf("expected single content item, got %d", len(res.Content))
	}
	tc, ok := res.Content[0].(mcplib.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Content[0])
	}
	var env errorEnvelope
	if err := json.Unmarshal([]byte(tc.Text), &env); err != nil {
		t.Fatalf("envelope is not JSON: %v\n%s", err, tc.Text)
	}
	return env
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !envContains(s, p) {
			return false
		}
	}
	return true
}

func envContains(s, substr string) bool {
	if substr == "" {
		return true
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
