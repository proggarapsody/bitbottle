package pr_test

// Bug: GetText sends no Accept header. Bitbucket Server returns JSON by
// default for the diff endpoint. The client must send Accept: text/plain
// to receive a unified diff.

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

// TestPRDiff_SendsAcceptTextPlain verifies that pr diff sends
// Accept: text/plain so Bitbucket Server returns a unified diff
// instead of its JSON structured-diff format.
func TestPRDiff_SendsAcceptTextPlain(t *testing.T) {
	t.Parallel()

	var gotAccept string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "diff --git a/foo b/foo\n+new\n")
	}))
	t.Cleanup(srv.Close)

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "ssh://git@bb.example.com:7999/myproj/myrepo.git\n"},
	)
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n",
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
		GitRunner:     runner,
	})

	root := pr.NewCmdPR(f)
	root.SetArgs([]string{"diff", "42"})
	require.NoError(t, root.Execute())

	assert.Equal(t, "text/plain", gotAccept,
		"pr diff must send Accept: text/plain so Bitbucket Server returns unified diff, not JSON")
}
