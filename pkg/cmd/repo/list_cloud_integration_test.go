package repo_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
)

// repoListCloudConfig wires bitbucket.org as the sole configured host so the
// command resolves it automatically without --hostname.
const repoListCloudConfig = "bitbucket.org:\n  oauth_token: tok\n  git_protocol: ssh\n"

// TestRepoList_CloudIntegration_PrintsSlugsFromCloudEnvelope verifies that
// `repo list` against a bitbucket.org host decodes the Cloud paged envelope
// ({pagelen,values,page,size}) and prints slug + namespace from full_name.
func TestRepoList_CloudIntegration_PrintsSlugsFromCloudEnvelope(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/cloud/testdata/repo_list.json")
	require.NoError(t, err)

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: repoListCloudConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	require.NoError(t, cmd.Execute())

	// Request went to the Cloud /repositories endpoint.
	assert.Contains(t, gotPath, "/repositories")

	output := out.String()
	assert.Contains(t, output, "my-service", "cloud slug should be printed")
	assert.Contains(t, output, "myworkspace", "cloud namespace should be printed")
}

// TestRepoList_CloudIntegration_ErrorContractSurfacesMessage verifies that a
// Cloud 401 error envelope surfaces the message text extracted from
// error.message when the command reports failure.
func TestRepoList_CloudIntegration_ErrorContractSurfacesMessage(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/cloud/testdata/error_401.json")
	require.NoError(t, err)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: repoListCloudConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
	assert.Contains(t, err.Error(), "Authentication required",
		"cloud error.message should propagate through backend.HTTPError")
}

// TestRepoList_CloudIntegration_EmptyCloudEnvelope verifies that the cloud
// empty list envelope yields no output and no error.
func TestRepoList_CloudIntegration_EmptyCloudEnvelope(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Cloud-shaped empty envelope (no isLastPage).
		_, _ = io.WriteString(w, `{"pagelen":10,"values":[],"page":1,"size":0}`)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: repoListCloudConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	require.NoError(t, cmd.Execute())
	assert.Empty(t, out.String())
}

// TestRepoList_CloudIntegration_BackendTypeOverrideForcesCloud verifies that
// a host with backend_type: cloud in hosts.yml, even though the hostname is
// not bitbucket.org, dispatches via the cloud adapter end-to-end (path is
// the Cloud shape /repositories).
func TestRepoList_CloudIntegration_BackendTypeOverrideForcesCloud(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"pagelen":10,"values":[],"page":1,"size":0}`)
	}))
	t.Cleanup(srv.Close)

	hostsYML := "internal.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n  backend_type: cloud\n"

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: hostsYML,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	require.NoError(t, cmd.Execute())

	assert.Contains(t, gotPath, "/repositories",
		"cloud backend_type override must route through cloud adapter")
}
