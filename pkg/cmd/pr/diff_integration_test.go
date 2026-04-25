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

// TestPRDiff_Integration_ServerEndToEnd exercises `pr diff ID` through the
// cobra command tree against a Bitbucket Data Center httptest server. Verifies
// the GET to /diff endpoint and that the raw diff text is streamed verbatim
// to stdout.
func TestPRDiff_Integration_ServerEndToEnd(t *testing.T) {
	t.Parallel()

	const diffBody = "diff --git a/foo b/foo\nindex 0000000..1111111 100644\n--- a/foo\n+++ b/foo\n@@ -1 +1 @@\n-old\n+new\n"

	var (
		gotMethod string
		gotPath   string
	)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, diffBody)
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
	root.SetArgs([]string{"diff", "42"})
	require.NoError(t, root.Execute())

	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Contains(t, gotPath, "/projects/MYPROJ/repos/myrepo/pull-requests/42/diff")
	assert.Equal(t, diffBody, out.String(), "raw diff should be streamed verbatim to stdout")
}
