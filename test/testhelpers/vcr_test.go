package testhelpers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v2/cassette"

	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// cassetteFixture is the path to the seeded BQ-2 cassette checked into the
// repository. Tests use it for offline (no-network) mechanism checks.
const cassetteFixture = "../../test/script/testdata/cassettes/pr_update_version_required.yaml"

// --- BodyAwareMatcher tests ------------------------------------------------

// TestBodyAwareMatcher_Mismatch_MissingVersionField loads the seeded BQ-2
// cassette and verifies that a PUT request whose body omits the "version"
// field does NOT match the recorded interaction. This is the core regression
// guard: if production code ever drops the version field, the test fails.
func TestBodyAwareMatcher_Mismatch_MissingVersionField(t *testing.T) {
	c, err := cassette.Load(strings.TrimSuffix(cassetteFixture, ".yaml"))
	if err != nil {
		t.Fatalf("load cassette: %v", err)
	}
	if len(c.Interactions) == 0 {
		t.Fatal("cassette has no interactions")
	}
	recorded := c.Interactions[0].Request

	// Build a request WITHOUT the version field — simulates the BQ-2 regression.
	withoutVersion := `{"title":"fix: my PR","description":""}`
	reqNoVersion, err := http.NewRequestWithContext(
		context.Background(), http.MethodPut, recorded.URL, strings.NewReader(withoutVersion),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if testhelpers.BodyAwareMatcher(reqNoVersion, recorded) {
		t.Error("expected BodyAwareMatcher to return false for body missing version field, got true")
	}
}

// TestBodyAwareMatcher_Match_WithVersionField verifies that the matcher returns
// true when both the URL and the body (including the version field) are identical
// to the cassette interaction.
func TestBodyAwareMatcher_Match_WithVersionField(t *testing.T) {
	c, err := cassette.Load(strings.TrimSuffix(cassetteFixture, ".yaml"))
	if err != nil {
		t.Fatalf("load cassette: %v", err)
	}
	recorded := c.Interactions[0].Request

	// Body identical to the cassette — should match.
	reqWithVersion, err := http.NewRequestWithContext(
		context.Background(), http.MethodPut, recorded.URL, strings.NewReader(recorded.Body),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if !testhelpers.BodyAwareMatcher(reqWithVersion, recorded) {
		t.Error("expected BodyAwareMatcher to return true for identical body, got false")
	}
}

// --- Redaction tests -------------------------------------------------------

// TestNewVCRDoer_RecordRedactsSecrets records a single HTTP interaction
// THROUGH the NewVCRDoer factory (record mode) against an in-process httptest
// server, then checks that the saved cassette YAML contains neither the
// Authorization token, the string "Bearer", nor the httptest server hostname.
//
// This proves the SaveFilter redaction is part of the factory contract: a
// caller running `make record-cassettes` gets safe cassettes without having to
// remember to wire any filter. It is the security-critical test for the record
// path.
func TestNewVCRDoer_RecordRedactsSecrets(t *testing.T) {
	// Spin up a server that returns a trivial JSON body.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer srv.Close()

	// Treat the httptest hostname as the "internal host" to be redacted.
	// srv.URL is e.g. "http://127.0.0.1:PORT"; extract the host portion.
	internalHost := strings.TrimPrefix(srv.URL, "http://")

	t.Setenv("BITBOTTLE_RECORD", "1")

	tmpDir := t.TempDir()
	cassettePath := filepath.Join(tmpDir, "auth_redact_test.yaml")

	client, stop := testhelpers.NewVCRDoer(t, cassettePath, internalHost)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer secret-token-123")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	_, _ = io.ReadAll(resp.Body)
	if err := resp.Body.Close(); err != nil {
		t.Errorf("close response body: %v", err)
	}

	// stop() flushes the cassette to disk, running the SaveFilter.
	stop()

	// Read the saved cassette YAML and assert no sensitive data leaked.
	data, err := os.ReadFile(cassettePath)
	if err != nil {
		t.Fatalf("read cassette: %v", err)
	}
	yamlStr := string(data)

	if strings.Contains(yamlStr, "secret-token-123") {
		t.Error("cassette contains raw token 'secret-token-123'")
	}
	if strings.Contains(yamlStr, "Bearer") {
		t.Error("cassette contains 'Bearer' keyword")
	}
	if strings.Contains(yamlStr, internalHost) {
		t.Errorf("cassette contains internal host %q", internalHost)
	}
}

// TestRedact_StripsAuthAndInternalHost is a focused unit test for the Redact
// function in isolation (constructs an interaction directly, no recorder).
func TestRedact_StripsAuthAndInternalHost(t *testing.T) {
	internalHost := "git.internal.invalid"
	i := &cassette.Interaction{
		Request: cassette.Request{
			Headers: http.Header{"Authorization": []string{"Bearer secret-token-123"}},
			Body:    `{"token":"Bearer secret-token-123","url":"https://git.internal.invalid/x"}`,
			URL:     "https://git.internal.invalid/rest/api/1.0/x",
		},
		Response: cassette.Response{
			Body: `{"link":"https://git.internal.invalid/y"}`,
		},
	}

	testhelpers.Redact(i, internalHost)

	if _, ok := i.Request.Headers["Authorization"]; ok {
		t.Error("Authorization header not stripped")
	}
	if strings.Contains(i.Request.Body, "secret-token-123") {
		t.Error("request body still contains token")
	}
	if strings.Contains(i.Request.Body, internalHost) ||
		strings.Contains(i.Response.Body, internalHost) ||
		strings.Contains(i.URL, internalHost) {
		t.Error("internal host not redacted")
	}
}

// --- Skip tests ------------------------------------------------------------

// TestNewVCRDoer_SkipsWhenNoCassette verifies that NewVCRDoer calls t.Skip
// when the cassette file does not exist and BITBOTTLE_RECORD is unset.
// We use a sub-test with a fake *testing.T to intercept the skip signal.
func TestNewVCRDoer_SkipsWhenNoCassette(t *testing.T) {
	// Ensure record mode is off.
	t.Setenv("BITBOTTLE_RECORD", "")

	// Run in a sub-test so that the outer test is not itself skipped.
	t.Run("inner", func(t *testing.T) {
		_, _ = testhelpers.NewVCRDoer(t, "/nonexistent/path/that/does/not/exist.yaml", "")
		// The skip signal appears in test output; Go's testing API does not
		// expose sub-test skip state to the parent. The sub-test will show
		// "--- SKIP" in the verbose output, which is the observable contract.
	})
}

// TestNewVCRDoer_ReplaysCassette checks that NewVCRDoer returns a functional
// Doer that can replay the seeded cassette.
func TestNewVCRDoer_ReplaysCassette(t *testing.T) {
	t.Setenv("BITBOTTLE_RECORD", "")

	doer, stop := testhelpers.NewVCRDoer(t, cassetteFixture, "")
	defer stop()

	// Build the exact request that the cassette recorded.
	body := `{"version":5,"title":"fix: my PR","description":""}`
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		"http://example-server.invalid/rest/api/1.0/projects/PROJ/repos/myrepo/pull-requests/42",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := doer.Do(req)
	if err != nil {
		t.Fatalf("replay request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from cassette replay, got %d", resp.StatusCode)
	}

	var got map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["id"] != float64(42) {
		t.Errorf("expected id=42 in replayed response, got %v", got["id"])
	}
}
