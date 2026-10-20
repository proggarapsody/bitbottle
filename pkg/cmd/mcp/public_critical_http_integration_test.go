package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	cmdmcp "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
)

const criticalHTTPChildEnv = "BITBOTTLE_TEST_MCP_CRITICAL_HTTP_CHILD"

// Only the selected child owns stdio; adapters use the parent's loopback fixture.
func TestMCPCriticalHTTPChild(t *testing.T) {
	if os.Getenv(criticalHTTPChildEnv) != "1" {
		return
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		ConfigDir:     os.Getenv("BITBOTTLE_TEST_MCP_CONFIG_DIR"),
		InitialConfig: "fixture.example:\n  oauth_token: test-token\n",
		BackendType:   os.Getenv("BITBOTTLE_TEST_MCP_PROVIDER"),
	})
	f.HTTPClient = factorytest.StubHTTPClient(&http.Client{Timeout: 5 * time.Second})
	f.BaseURL = func(string) string { return os.Getenv("BITBOTTLE_TEST_MCP_BASE_URL") }
	cmd := cmdmcp.NewCmdMCP(f)
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func startCriticalHTTPSession(t *testing.T, provider, baseURL string) *stdioSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	t.Cleanup(cancel)
	configDir := t.TempDir()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMCPCriticalHTTPChild$")
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + configDir, "TMPDIR=" + configDir, criticalHTTPChildEnv + "=1", "BITBOTTLE_TEST_MCP_CONFIG_DIR=" + configDir, "BITBOTTLE_TEST_MCP_PROVIDER=" + provider, "BITBOTTLE_TEST_MCP_BASE_URL=" + baseURL}
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())
	s := &stdioSession{ctx: ctx, cmd: cmd, stdin: stdin, relayDone: make(chan error, 1), rawStdout: &bytes.Buffer{}, stderr: &stderr}
	t.Cleanup(func() {
		_ = stdin.Close()
		if !s.waited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	input, output := io.Pipe()
	s.clientInput = input
	s.relayOutput = output
	go func() {
		_, copyErr := io.Copy(io.MultiWriter(output, s.rawStdout), stdout)
		_ = output.CloseWithError(copyErr)
		s.relayDone <- copyErr
	}()
	protocol := client.NewClient(transport.NewIO(input, stdin, io.NopCloser(strings.NewReader(""))))
	s.client = protocol
	t.Cleanup(func() { _ = protocol.Close(); _ = input.Close(); _ = output.Close() })
	require.NoError(t, protocol.Start(ctx))
	s.initialized, err = protocol.Initialize(ctx, mcplib.InitializeRequest{Params: mcplib.InitializeParams{ProtocolVersion: mcplib.LATEST_PROTOCOL_VERSION, ClientInfo: mcplib.Implementation{Name: "bitbottle-core-matrix-test", Version: "1"}}})
	require.NoError(t, err)
	return s
}

func closeCriticalHTTPSession(t *testing.T, s *stdioSession) {
	t.Helper()
	require.NoError(t, s.stdin.Close())
	select {
	case err := <-s.relayDone:
		require.NoError(t, err)
	case <-s.ctx.Done():
		t.Fatalf("MCP stdout did not close: %v", s.ctx.Err())
	}
	require.NoError(t, s.cmd.Wait(), "MCP child stderr: %s", s.stderr.String())
	s.waited = true
	for _, line := range bytes.Split(bytes.TrimSpace(s.rawStdout.Bytes()), []byte{'\n'}) {
		assert.True(t, json.Valid(line), "stdout must contain JSON-RPC, got %s", line)
	}
}

// These are external HTTP contract fixtures, not live-provider acceptance.
// Comment examples match api/{cloud,server}/pr_comments_test.go; Server's
// nonzero version precondition is also documented in docs/backend-quirks.md.
type criticalHTTPRequest struct {
	method, uri, body string
	status            int
	response          string
}

func criticalHTTPFixture(t *testing.T, requests ...criticalHTTPRequest) *httptest.Server {
	t.Helper()
	steps := make(chan criticalHTTPRequest, len(requests))
	for _, request := range requests {
		steps <- request
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case want := <-steps:
			assert.Equal(t, want.method, r.Method, "provider request method")
			assert.Equal(t, want.uri, r.URL.RequestURI(), "provider request URI")
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			if want.body != "" {
				assert.JSONEq(t, want.body, string(body), "provider request payload")
			} else {
				assert.Empty(t, string(body))
			}
			if want.body != "" || (r.Method != http.MethodGet && strings.HasPrefix(want.uri, "/rest/")) {
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			}
			w.Header().Set("Content-Type", "application/json")
			status := want.status
			if status == 0 {
				status = http.StatusOK
			}
			w.WriteHeader(status)
			_, _ = io.WriteString(w, want.response)
		default:
			t.Errorf("unexpected provider request: %s %s", r.Method, r.URL.RequestURI())
			http.Error(w, "unexpected request", http.StatusInternalServerError)
		}
	}))
	t.Cleanup(func() { srv.Close(); assert.Empty(t, steps, "expected provider requests were not made") })
	return srv
}

func criticalHTTPCall(t *testing.T, session *stdioSession, name string, extra map[string]any) *mcplib.CallToolResult {
	t.Helper()
	args := map[string]any{"hostname": "fixture.example", "project": "TEAM", "slug": "service", "id": 42}
	for key, value := range extra {
		args[key] = value
	}
	result, err := session.call(name, args)
	require.NoError(t, err)
	require.NotNil(t, result)
	return result
}

func criticalHTTPComment(t *testing.T, result *mcplib.CallToolResult) backend.PRComment {
	t.Helper()
	require.False(t, result.IsError, resultText(t, result))
	var comment backend.PRComment
	require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &comment))
	return comment
}

func TestMCPPublicCriticalHTTP_CloudCommentLifecycle(t *testing.T) {
	const collection = "/repositories/TEAM/service/pullrequests/42/comments"
	srv := criticalHTTPFixture(t,
		criticalHTTPRequest{method: "POST", uri: collection, body: `{"content":{"raw":"review this"},"inline":{"path":"main.go","to":42}}`, response: `{"id":7,"content":{"raw":"review this"},"user":{"nickname":"alice"},"inline":{"path":"main.go","to":42}}`},
		criticalHTTPRequest{method: "POST", uri: collection, body: `{"content":{"raw":"reply"},"parent":{"id":7}}`, response: `{"id":8,"content":{"raw":"reply"},"parent":{"id":7}}`},
		criticalHTTPRequest{method: "PUT", uri: collection + "/8", body: `{"content":{"raw":"edited reply"}}`, response: `{"id":8,"content":{"raw":"edited reply"},"parent":{"id":7}}`},
		criticalHTTPRequest{method: "DELETE", uri: collection + "/8", status: 204},
	)
	s := startCriticalHTTPSession(t, "cloud", srv.URL)
	inline := criticalHTTPComment(t, criticalHTTPCall(t, s, "add_pr_comment", map[string]any{"body": "review this", "inline_path": "main.go", "inline_line": 42, "inline_side": "new"}))
	assert.Equal(t, 7, inline.ID)
	assert.Equal(t, "review this", inline.Text)
	assert.Equal(t, "alice", inline.Author.Slug)
	require.NotNil(t, inline.Inline)
	assert.Equal(t, backend.PRCommentInline{Path: "main.go", Side: "new", Line: 42}, *inline.Inline)
	reply := criticalHTTPComment(t, criticalHTTPCall(t, s, "add_pr_comment", map[string]any{"body": "reply", "parent_id": 7}))
	assert.Equal(t, 8, reply.ID)
	assert.Equal(t, 7, reply.ParentID)
	edited := criticalHTTPComment(t, criticalHTTPCall(t, s, "edit_pr_comment", map[string]any{"body": "edited reply", "comment_id": 8}))
	assert.Equal(t, "edited reply", edited.Text)
	assert.Equal(t, 7, edited.ParentID)
	deleted := criticalHTTPCall(t, s, "delete_pr_comment", map[string]any{"comment_id": 8})
	require.False(t, deleted.IsError, resultText(t, deleted))
	assert.JSONEq(t, `{}`, resultText(t, deleted))
	closeCriticalHTTPSession(t, s)
}

func TestMCPPublicCriticalHTTP_ServerCommentConcurrency(t *testing.T) {
	const comment = "/rest/api/1.0/projects/TEAM/repos/service/pull-requests/42/comments/99"
	t.Run("edit and delete use independently fetched versions", func(t *testing.T) {
		srv := criticalHTTPFixture(t,
			criticalHTTPRequest{method: "GET", uri: comment, response: `{"id":99,"version":3,"text":"old"}`},
			criticalHTTPRequest{method: "PUT", uri: comment + "?version=3", body: `{"text":"new","version":3}`, response: `{"id":99,"version":4,"text":"new","author":{"slug":"alice"}}`},
			criticalHTTPRequest{method: "GET", uri: comment, response: `{"id":99,"version":5,"text":"new"}`},
			criticalHTTPRequest{method: "DELETE", uri: comment + "?version=5", status: 204},
		)
		s := startCriticalHTTPSession(t, "server", srv.URL+"/rest/api/1.0")
		edited := criticalHTTPComment(t, criticalHTTPCall(t, s, "edit_pr_comment", map[string]any{"body": "new", "comment_id": 99}))
		assert.Equal(t, 99, edited.ID)
		assert.Equal(t, "new", edited.Text)
		assert.Equal(t, 4, edited.Version)
		assert.Equal(t, "alice", edited.Author.Slug)
		deleted := criticalHTTPCall(t, s, "delete_pr_comment", map[string]any{"comment_id": 99})
		require.False(t, deleted.IsError, resultText(t, deleted))
		assert.JSONEq(t, `{}`, resultText(t, deleted))
		closeCriticalHTTPSession(t, s)
	})
	for _, operation := range []string{"edit_pr_comment", "delete_pr_comment"} {
		t.Run(operation+" stops when version read is forbidden", func(t *testing.T) {
			srv := criticalHTTPFixture(t, criticalHTTPRequest{method: "GET", uri: comment, status: 403, response: `{"errors":[{"message":"comment access denied"}]}`})
			s := startCriticalHTTPSession(t, "server", srv.URL+"/rest/api/1.0")
			result := criticalHTTPCall(t, s, operation, map[string]any{"body": "must not write", "comment_id": 99})
			require.True(t, result.IsError, resultText(t, result))
			assert.Contains(t, resultText(t, result), `"code":"permission"`)
			assert.Contains(t, resultText(t, result), "comment access denied")
			closeCriticalHTTPSession(t, s)
		})
	}
}

func TestMCPPublicCriticalHTTP_OrdinaryMerge(t *testing.T) {
	const cloudPR = "/repositories/TEAM/service/pullrequests/42"
	// Strategy literals are the provider examples in api/{cloud,server}/prs_test.go.
	t.Run("Cloud maps strategy and returns confirmed merge", func(t *testing.T) {
		srv := criticalHTTPFixture(t, criticalHTTPRequest{method: "POST", uri: cloudPR + "/merge", body: `{"merge_strategy":"rebase_fast_forward"}`, response: `{"id":42,"state":"MERGED","title":"Merged remotely"}`})
		s := startCriticalHTTPSession(t, "cloud", srv.URL)
		result := criticalHTTPCall(t, s, "merge_pr", map[string]any{"strategy": "rebase"})
		require.False(t, result.IsError, resultText(t, result))
		var pr backend.PullRequest
		require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &pr))
		assert.Equal(t, 42, pr.ID)
		assert.Equal(t, "MERGED", pr.State)
		assert.Equal(t, "Merged remotely", pr.Title)
		closeCriticalHTTPSession(t, s)
	})
	for _, empty := range []bool{false, true} {
		name := "OPEN response is not merged"
		response := `{"id":42,"state":"OPEN"}`
		if empty {
			name = "empty accepted response requires readback"
			response = ""
		}
		t.Run("Cloud "+name, func(t *testing.T) {
			requests := []criticalHTTPRequest{{method: "POST", uri: cloudPR + "/merge", body: `{"merge_strategy":"merge_commit"}`, status: 202, response: response}}
			if empty {
				requests = append(requests, criticalHTTPRequest{method: "GET", uri: cloudPR, response: `{"id":42,"state":"OPEN"}`})
			}
			srv := criticalHTTPFixture(t, requests...)
			s := startCriticalHTTPSession(t, "cloud", srv.URL)
			result := criticalHTTPCall(t, s, "merge_pr", map[string]any{"strategy": "merge"})
			require.True(t, result.IsError, resultText(t, result))
			assert.Contains(t, resultText(t, result), "pr.merge.pending")
			closeCriticalHTTPSession(t, s)
		})
	}
	t.Run("Server preserves current version and strategy", func(t *testing.T) {
		const prPath = "/rest/api/1.0/projects/TEAM/repos/service/pull-requests/42"
		srv := criticalHTTPFixture(t,
			criticalHTTPRequest{method: "GET", uri: prPath, response: `{"id":42,"version":17,"state":"OPEN"}`},
			criticalHTTPRequest{method: "POST", uri: prPath + "/merge", body: `{"version":17,"strategyId":"rebase-ff-only"}`, response: `{"id":42,"version":18,"state":"MERGED","title":"Server merged"}`},
		)
		s := startCriticalHTTPSession(t, "server", srv.URL+"/rest/api/1.0")
		result := criticalHTTPCall(t, s, "merge_pr", map[string]any{"strategy": "rebase"})
		require.False(t, result.IsError, resultText(t, result))
		var pr backend.PullRequest
		require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &pr))
		assert.Equal(t, 42, pr.ID)
		assert.Equal(t, "MERGED", pr.State)
		assert.Equal(t, 18, pr.Version)
		closeCriticalHTTPSession(t, s)
	})
}
