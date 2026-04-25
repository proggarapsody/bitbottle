package repo_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
)

// TestRepoView_Integration_ServerEndToEnd exercises `repo view PROJECT/REPO`
// against a Bitbucket Data Center httptest server. Verifies the GET path,
// JSON decoding, and that the rendered output contains the slug, namespace,
// SCM, and web URL.
func TestRepoView_Integration_ServerEndToEnd(t *testing.T) {
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
			"id": 7,
			"slug": "myrepo",
			"name": "myrepo",
			"project": {"key": "MYPROJ", "name": "My Project"},
			"scmId": "git",
			"state": "AVAILABLE",
			"links": {"self": [{"href": "https://bb.example.com/projects/MYPROJ/repos/myrepo/browse"}]}
		}`)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n",
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	root := repo.NewCmdRepo(f)
	root.SetArgs([]string{"view", "MYPROJ/myrepo"})
	require.NoError(t, root.Execute())

	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Contains(t, gotPath, "/projects/MYPROJ/repos/myrepo")

	output := out.String()
	assert.Contains(t, output, "myrepo", "slug should be rendered")
	assert.Contains(t, output, "MYPROJ", "namespace should be rendered")
	assert.Contains(t, output, "git", "SCM should be rendered")
	assert.Contains(t, output, "browse", "web URL should be rendered")
}
