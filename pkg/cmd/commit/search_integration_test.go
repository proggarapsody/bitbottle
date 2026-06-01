package commit_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
)

// serverCommitSearchConfig targets a Bitbucket Server / Data Center host so
// the factory wires a server.Client and hits the correct REST path.
const serverCommitSearchConfig = "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n"

// searchCommitsJSON is a minimal Bitbucket Server paged commit response.
const searchCommitsJSON = `{"values":[` +
	`{"id":"abc1234def456abc1234def456abc1234def456ab","message":"feat: add search endpoint","author":{"name":"alice","emailAddress":"alice@example.com"},"authorTimestamp":1714118400000},` +
	`{"id":"bbb2234def456bbb2234def456bbb2234def456bb","message":"fix: null pointer","author":{"name":"bob","emailAddress":"bob@example.com"},"authorTimestamp":1714032000000}` +
	`],"isLastPage":true}`

// TestCommitSearch_Integration_PrintsTable verifies that `commit search PROJECT/REPO`
// fetches commits over HTTP and renders them in table format.
func TestCommitSearch_Integration_PrintsTable(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsJSON))
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverCommitSearchConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(hostname string) string { return srv.URL }

	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "abc1234d") // 8-char hash truncation in non-TTY
	assert.Contains(t, output, "feat: add search endpoint")
	assert.Contains(t, output, "alice")
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/commits", gotPath)
}

// TestCommitSearch_Integration_JSONOutput verifies that --json produces valid
// JSON containing all commit hashes.
func TestCommitSearch_Integration_JSONOutput(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsJSON))
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverCommitSearchConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(hostname string) string { return srv.URL }

	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	var results []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &results))
	require.Len(t, results, 2)
	assert.Equal(t, "abc1234def456abc1234def456abc1234def456ab", results[0]["hash"])
	assert.Equal(t, "bbb2234def456bbb2234def456bbb2234def456bb", results[1]["hash"])
}

// TestCommitSearch_Integration_QueryFilterPassedToServer verifies that --query
// does not appear as a URL param (Server filters client-side) and that the
// result set is filtered: only the commit whose message matches is returned.
func TestCommitSearch_Integration_QueryFilterPassedToServer(t *testing.T) {
	t.Parallel()

	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsJSON))
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverCommitSearchConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(hostname string) string { return srv.URL }

	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--query", "search"})
	require.NoError(t, cmd.Execute())

	// Server does not accept a q= parameter; filter is client-side.
	assert.Empty(t, gotQuery)
	// Only "feat: add search endpoint" contains "search".
	output := out.String()
	assert.Contains(t, output, "feat: add search endpoint")
	assert.NotContains(t, output, "fix: null pointer")
}

// TestCommitSearch_Integration_APIErrorSurfaced verifies that an HTTP 401
// causes the command to return an error.
func TestCommitSearch_Integration_APIErrorSurfaced(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"errors":[{"message":"Unauthenticated"}]}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverCommitSearchConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(hostname string) string { return srv.URL }

	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
}

// TestCommitSearch_Integration_EmptyResultPrintsNothing verifies that an empty
// response produces no output and exits cleanly.
func TestCommitSearch_Integration_EmptyResultPrintsNothing(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverCommitSearchConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(hostname string) string { return srv.URL }

	cmd := commit.NewCmdCommitSearch(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Empty(t, out.String())
}
