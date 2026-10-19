package pr_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestPRMerge_Integration_ServerEndToEnd exercises `pr merge ID --merge` through
// the cobra command tree against a Bitbucket Data Center httptest server.
// Verifies the merge is POSTed to the /merge subpath and that the user-visible
// "Merged pull request" line is written to stdout.
func TestPRMerge_Integration_ServerEndToEnd(t *testing.T) {
	t.Parallel()

	var (
		gotMethod string
		gotPath   string
	)
	prJSON := `{
		"id": 42,
		"version": 1,
		"title": "Done",
		"state": "MERGED",
		"author": {"user": {"slug": "alice", "displayName": "Alice"}},
		"fromRef": {"id": "refs/heads/feat/x", "displayId": "feat/x"},
		"toRef": {"id": "refs/heads/main", "displayId": "main"}
	}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record only the merge POST — the GET that fetches the version fires first.
		if r.Method == http.MethodPost {
			gotMethod = r.Method
			gotPath = r.URL.Path
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, prJSON)
	}))
	t.Cleanup(srv.Close)

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "ssh://git@bb.example.com:7999/myproj/myrepo.git\n"},
	)

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n"})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(hostname string) string { return srv.URL }
	f.GitRunner = func() run.Runner { return runner }

	root := pr.NewCmdPR(f)
	root.SetArgs([]string{"merge", "42", "--merge"})
	require.NoError(t, root.Execute())

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Contains(t, gotPath, "/projects/MYPROJ/repos/myrepo/pull-requests/42/merge")
	assert.Contains(t, out.String(), "Merged pull request")
	assert.Contains(t, out.String(), "#42")
}

func TestPRMerge_Integration_RebaseUsesProviderStrategy(t *testing.T) {
	for _, tc := range []struct {
		name, config, remote, expectedPath, expectedStrategy string
	}{
		{"cloud", "bitbucket.org:\n  oauth_token: tok\n  git_protocol: ssh\n", "git@bitbucket.org:myworkspace/myrepo.git\n", "/repositories/MYWORKSPACE/myrepo/pullrequests/42/merge", "rebase_fast_forward"},
		{"data-center", "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n", "ssh://git@bb.example.com:7999/myproj/myrepo.git\n", "/projects/MYPROJ/repos/myrepo/pull-requests/42/merge", "rebase-ff-only"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotPath string
			var gotBody map[string]any
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPost {
					gotPath = r.URL.Path
					body, _ := io.ReadAll(r.Body)
					_ = json.Unmarshal(body, &gotBody)
				}
				_, _ = io.WriteString(w, `{"id":42,"version":3,"state":"MERGED"}`)
			}))
			t.Cleanup(srv.Close)

			runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{Stdout: tc.remote})
			f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: tc.config})
			f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
			f.BaseURL = func(hostname string) string { return srv.URL }
			f.GitRunner = func() run.Runner { return runner }

			root := pr.NewCmdPR(f)
			root.SetArgs([]string{"merge", "42", "--rebase"})
			require.NoError(t, root.Execute())
			assert.Equal(t, tc.expectedPath, gotPath)
			if tc.name == "cloud" {
				assert.Equal(t, tc.expectedStrategy, gotBody["merge_strategy"])
			} else {
				assert.Equal(t, tc.expectedStrategy, gotBody["strategyId"])
			}
			assert.Contains(t, out.String(), "Merged pull request #42")
		})
	}
}
