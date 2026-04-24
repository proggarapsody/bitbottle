package pr_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
)

const prListCloudConfig = "bitbucket.org:\n  oauth_token: tok\n  git_protocol: ssh\n"

// TestPRList_CloudIntegration_PrintsTitlesFromCloudEnvelope verifies that
// `pr list workspace/my-svc` against bitbucket.org hits the Cloud
// /repositories/.../pullrequests endpoint and prints the title.
func TestPRList_CloudIntegration_PrintsTitlesFromCloudEnvelope(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/cloud/testdata/pr_list.json")
	require.NoError(t, err)

	var gotPath, gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListCloudConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	// Cloud path shape.
	assert.Contains(t, gotPath, "/repositories/myworkspace/my-service/pullrequests")
	// Cloud uses state=OPEN and pagelen= query string.
	assert.Contains(t, gotQuery, "state=OPEN")
	assert.Contains(t, gotQuery, "pagelen=")

	output := out.String()
	assert.Contains(t, output, "Fix login bug")
}

// TestPRList_CloudIntegration_EmptyCloudEnvelope verifies that an empty
// Cloud PR list produces no output.
func TestPRList_CloudIntegration_EmptyCloudEnvelope(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"pagelen":10,"values":[],"page":1,"size":0}`)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListCloudConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Empty(t, out.String())
}

// TestPRList_CloudIntegration_ErrorContractSurfacesMessage verifies that a
// Cloud 404 error envelope surfaces the message.
func TestPRList_CloudIntegration_ErrorContractSurfacesMessage(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/cloud/testdata/error_404.json")
	require.NoError(t, err)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prListCloudConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"myworkspace/missing"})
	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

// TestPRList_CloudIntegration_BackendTypeOverrideForcesCloud verifies that
// `pr list` on a host configured with backend_type: cloud routes through the
// Cloud adapter (path is /repositories/.../pullrequests, not /rest/api/...).
func TestPRList_CloudIntegration_BackendTypeOverrideForcesCloud(t *testing.T) {
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

	cmd := pr.NewCmdPRList(f)
	// Explicit --hostname so the command uses the configured host.
	cmd.SetArgs([]string{"ws/svc", "--hostname", "internal.example.com"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, gotPath, "/repositories/ws/svc/pullrequests",
		"cloud backend_type override must route through cloud adapter")
}
