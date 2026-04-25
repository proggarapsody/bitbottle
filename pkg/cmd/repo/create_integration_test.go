package repo_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
)

// TestRepoCreate_Integration_ServerEndToEnd wires the full `repo create` command
// through cobra to a Bitbucket Data Center httptest server. It verifies:
//   - the command issues POST /rest/api/1.0/projects/MYPROJ/repos
//   - the request body contains the supplied repo name and SCM
//   - the response is decoded and the user-visible "Created repository" line
//     reaches stdout.
func TestRepoCreate_Integration_ServerEndToEnd(t *testing.T) {
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

	// Build the full repo command tree so we exercise cobra dispatch end to end.
	root := repo.NewCmdRepo(f)
	root.SetArgs([]string{"create", "myrepo", "--project", "MYPROJ"})
	require.NoError(t, root.Execute())

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Contains(t, gotPath, "/projects/MYPROJ/repos")
	assert.Equal(t, "myrepo", gotBody["name"], "request body must contain repo name")
	assert.Equal(t, "git", gotBody["scmId"])
	assert.Contains(t, out.String(), "Created repository")
	assert.Contains(t, out.String(), "myrepo")
}

// TestRepoCreate_Integration_ServerError verifies a 409 from the server surfaces
// as a command-level error (cobra returns the wrapped HTTPError).
func TestRepoCreate_Integration_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"errors":[{"message":"already exists"}]}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n",
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	root := repo.NewCmdRepo(f)
	root.SetArgs([]string{"create", "myrepo", "--project", "MYPROJ"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "409")
}
