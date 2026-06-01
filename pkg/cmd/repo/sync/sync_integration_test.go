package sync_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/sync"
)

// cloudConfig wires bitbucket.org as the sole configured host so the command
// resolves it automatically without --hostname.
const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  git_protocol: ssh\n"

// serverConfig wires a Server / Data Center host so AsRepoSyncer returns
// ErrUnsupportedOnHost.
const serverConfig = "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n"

// TestRepoSync_Integration_SuccessPath verifies that a successful Cloud
// merge-upstream response prints the human-readable "Synced N commit(s)" line.
func TestRepoSync_Integration_SuccessPath(t *testing.T) {
	t.Parallel()

	var gotMethod, gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"behind":3,"commits_merged":3}`)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(_ string) string { return srv.URL }

	cmd := sync.NewCmdSync(f)
	cmd.SetArgs([]string{"myworkspace/my-fork"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Contains(t, gotPath, "/repositories/myworkspace/my-fork/merge-upstream")
	assert.Contains(t, out.String(), "Synced 3 commit(s) from upstream")
}

// TestRepoSync_Integration_AlreadyUpToDate verifies the zero-commits-merged branch.
func TestRepoSync_Integration_AlreadyUpToDate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"behind":0,"commits_merged":0}`)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(_ string) string { return srv.URL }

	cmd := sync.NewCmdSync(f)
	cmd.SetArgs([]string{"myworkspace/my-fork"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "Already up to date")
}

// TestRepoSync_Integration_JSONOutput verifies --json emits the structured result.
func TestRepoSync_Integration_JSONOutput(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"behind":2,"commits_merged":2}`)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(_ string) string { return srv.URL }

	cmd := sync.NewCmdSync(f)
	cmd.SetArgs([]string{"myworkspace/my-fork", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"behind"`)
	assert.Contains(t, got, `"commits_merged"`)
}

// TestRepoSync_Integration_UnsupportedOnServer verifies that running sync
// against a Bitbucket Server / Data Center host surfaces ErrUnsupportedOnHost.
func TestRepoSync_Integration_UnsupportedOnServer(t *testing.T) {
	t.Parallel()

	// The server stub should never be reached for the sync request because
	// AsRepoSyncer fails before any HTTP call is made. We still need a live
	// TLS server so the factory can resolve the backend without error.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(_ string) string { return srv.URL }

	cmd := sync.NewCmdSync(f)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--hostname", "bb.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}
