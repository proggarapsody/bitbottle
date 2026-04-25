package pr_test

import (
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

// TestPRView_Integration_ServerEndToEnd exercises `pr view ID` through the
// cobra command tree against a Bitbucket Data Center httptest server. Verifies
// the GET path and that the rendered output contains the PR title, author,
// and branches.
func TestPRView_Integration_ServerEndToEnd(t *testing.T) {
	t.Parallel()

	var (
		gotMethod string
		gotPath   string
	)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"id": 42,
			"title": "Hello PR",
			"description": "body",
			"state": "OPEN",
			"draft": false,
			"author": {"user": {"slug": "alice", "displayName": "Alice"}},
			"fromRef": {"id": "refs/heads/feat/x", "displayId": "feat/x"},
			"toRef": {"id": "refs/heads/main", "displayId": "main"},
			"links": {"self": [{"href": "https://bb.example.com/projects/MYPROJ/repos/myrepo/pull-requests/42"}]}
		}`)
	}))
	t.Cleanup(srv.Close)

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "ssh://git@bb.example.com:7999/myproj/myrepo.git\n"},
	)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n",
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
		GitRunner:     runner,
	})

	root := pr.NewCmdPR(f)
	root.SetArgs([]string{"view", "42"})
	require.NoError(t, root.Execute())

	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Contains(t, gotPath, "/projects/MYPROJ/repos/myrepo/pull-requests/42")

	output := out.String()
	assert.Contains(t, output, "Hello PR", "title should be in output")
	assert.Contains(t, output, "alice", "author slug should be in output")
	assert.Contains(t, output, "feat/x", "from branch should be in output")
	assert.Contains(t, output, "main", "to branch should be in output")
}
