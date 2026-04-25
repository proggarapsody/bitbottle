package pr_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestPRApprove_Integration_ServerEndToEnd exercises `pr approve ID` through
// the cobra command tree against a Bitbucket Data Center httptest server.
// Verifies the PUT to /participants/~ endpoint and the success message.
func TestPRApprove_Integration_ServerEndToEnd(t *testing.T) {
	t.Parallel()

	var (
		gotMethod string
		gotPath   string
	)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
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
	root.SetArgs([]string{"approve", "42"})
	require.NoError(t, root.Execute())

	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Contains(t, gotPath, "/projects/MYPROJ/repos/myrepo/pull-requests/42/participants")
	assert.Contains(t, out.String(), "Approved pull request")
	assert.Contains(t, out.String(), "#42")
}
