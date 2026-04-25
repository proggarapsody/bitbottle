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

// TestRepoDelete_Integration_ServerEndToEnd exercises `repo delete --confirm`
// against a Bitbucket Data Center httptest server. The bypass of the TTY
// confirmation prompt with --confirm keeps this test purely about HTTP wiring.
func TestRepoDelete_Integration_ServerEndToEnd(t *testing.T) {
	t.Parallel()

	var (
		gotMethod string
		gotPath   string
	)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n",
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	root := repo.NewCmdRepo(f)
	root.SetArgs([]string{"delete", "MYPROJ/myrepo", "--confirm"})
	require.NoError(t, root.Execute())

	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Contains(t, gotPath, "/projects/MYPROJ/repos/myrepo")
	assert.Contains(t, out.String(), "Deleted repository")
	assert.Contains(t, out.String(), "MYPROJ/myrepo")
}

// TestRepoDelete_Integration_ServerError verifies a 404 from the server surfaces
// as a command-level error.
func TestRepoDelete_Integration_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"repo not found"}]}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n",
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	root := repo.NewCmdRepo(f)
	root.SetArgs([]string{"delete", "MYPROJ/myrepo", "--confirm"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
