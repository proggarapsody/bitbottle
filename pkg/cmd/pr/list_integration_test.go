package pr_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/aleksey/bitbottle/pkg/cmd/pr"
	"github.com/aleksey/bitbottle/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const prListConfig = "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n"

// TestPRList_Integration_PrintsPRTitles verifies that `pr list PROJ/REPO`
// fetches open PRs and writes titles to stdout.
func TestPRList_Integration_PrintsPRTitles(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/testdata/pr_list.json")
	require.NoError(t, err)

	var gotState string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotState = r.URL.Query().Get("state")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "Fix login bug")
	assert.Contains(t, output, "Add dark mode")
	assert.Equal(t, "OPEN", gotState)
}

// TestPRList_Integration_StateFilterPassedToAPI verifies that --state=merged
// is forwarded as the `state` query parameter (uppercased).
func TestPRList_Integration_StateFilterPassedToAPI(t *testing.T) {
	t.Parallel()

	var gotState string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotState = r.URL.Query().Get("state")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--state", "merged"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MERGED", gotState)
}

// TestPRList_Integration_APIErrorSurfaced verifies that an API 401 causes the
// command to return an error.
func TestPRList_Integration_APIErrorSurfaced(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/testdata/error_401.json")
	require.NoError(t, err)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err = cmd.Execute()
	require.Error(t, err)
}

// TestPRList_Integration_EmptyResultPrintsNothing verifies that an empty PR
// list produces no output and exits cleanly.
func TestPRList_Integration_EmptyResultPrintsNothing(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Empty(t, out.String())
}

// TestPRList_Integration_DetectsFromGitRemote verifies that when no PROJECT/REPO
// argument is provided the command detects project and slug from the git remote.
func TestPRList_Integration_DetectsFromGitRemote(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "git@bb.example.com:MYPROJ/my-service.git\n"},
	)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
		GitRunner:     runner,
	})

	cmd := pr.NewCmdPRList(f)
	require.NoError(t, cmd.Execute())

	assert.Contains(t, gotPath, "MYPROJ")
	assert.Contains(t, gotPath, "my-service")
}

// TestPRList_Integration_ServerError verifies that a 5xx response causes the
// command to return an error containing the HTTP status code.
func TestPRList_Integration_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"errors":[{"message":"internal server error"}]}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
