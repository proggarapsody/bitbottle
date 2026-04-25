package pr_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestPRCreate_Integration_ServerEndToEnd exercises `pr create` end-to-end:
// the command resolves the repo from a fake git remote, calls `git rev-parse`
// for the current branch, then POSTs to the Bitbucket Data Center pull-requests
// endpoint. We verify the HTTP request shape and the user-visible output.
func TestPRCreate_Integration_ServerEndToEnd(t *testing.T) {
	t.Parallel()

	var (
		gotMethod string
		gotPath   string
		gotBody   map[string]any
	)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{
			"id": 101,
			"title": "Add feature X",
			"state": "OPEN",
			"draft": false,
			"author": {"user": {"slug": "alice", "displayName": "Alice"}},
			"fromRef": {"id": "refs/heads/feat/x", "displayId": "feat/x"},
			"toRef": {"id": "refs/heads/main", "displayId": "main"},
			"links": {"self": [{"href": "https://bb.example.com/projects/MYPROJ/repos/myrepo/pull-requests/101"}]}
		}`)
	}))
	t.Cleanup(srv.Close)

	// FakeRunner: first call returns the origin remote URL, second returns the
	// current branch (rev-parse). Order matches resolveRepoRef → CurrentBranch.
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "ssh://git@bb.example.com:7999/myproj/myrepo.git\n"},
		testhelpers.RunResponse{Stdout: "feat/x\n"},
	)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n",
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
		GitRunner:     runner,
	})

	root := pr.NewCmdPR(f)
	root.SetArgs([]string{"create", "--title", "Add feature X", "--base", "main"})
	require.NoError(t, root.Execute())

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Contains(t, gotPath, "/projects/MYPROJ/repos/myrepo/pull-requests")
	assert.Equal(t, "Add feature X", gotBody["title"])
	assert.Contains(t, out.String(), "Created pull request")
	assert.Contains(t, out.String(), "Add feature X")
}
