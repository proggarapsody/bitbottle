package view_test

// Bug: repo view lacks --hostname flag, making it unusable when multiple
// hosts are configured (unlike every other view/list command).

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/view"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoView_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := view.NewCmdView(f)
	assert.NotNil(t, cmd.Flag("hostname"), "repo view should have a --hostname flag like repo list")
}

func TestRepoView_ExplicitHostname_UsesProvidedHost(t *testing.T) {
	t.Parallel()

	var gotHost string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"slug":    "my-service",
			"name":    "My Service",
			"scmId":   "git",
			"project": map[string]any{"key": "MYPROJ"},
		})
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "bb1.example.com:\n  oauth_token: tok1\nbb2.example.com:\n  oauth_token: tok2\n"})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(hostname string) string {
		gotHost = hostname
		return srv.URL
	}

	cmd := view.NewCmdView(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--hostname", "bb2.example.com"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "bb2.example.com", gotHost)
}

func TestRepoView_MultipleHosts_NoFlag_ReturnsError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return testhelpers.BackendRepoFactory(), nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "bb1.example.com:\n  oauth_token: tok1\nbb2.example.com:\n  oauth_token: tok2\n"})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple hosts")
}
