package cloud_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const listRepoWatchersJSON = `{"values":[{"display_name":"Alice Smith","username":"alice"},{"display_name":"Bob Jones","username":"bob"}]}`

func TestCloudClient_ListRepoWatchers_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.ListRepoWatchers("myworkspace", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/watchers", gotPath)
}

func TestCloudClient_ListRepoWatchers_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listRepoWatchersJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	watchers, err := client.ListRepoWatchers("myworkspace", "my-service")
	require.NoError(t, err)
	require.Len(t, watchers, 2)
	assert.Equal(t, "alice", watchers[0].Slug)
	assert.Equal(t, "Alice Smith", watchers[0].DisplayName)
	assert.Equal(t, "bob", watchers[1].Slug)
	assert.Equal(t, "Bob Jones", watchers[1].DisplayName)
}

func TestCloudClient_ListRepoWatchers_FollowsPagination(t *testing.T) {
	t.Parallel()
	page1Tmpl := `{"values":[{"display_name":"Alice","username":"alice"}],"next":"%s/page2"}`
	page2 := `{"values":[{"display_name":"Bob","username":"bob"}]}`

	var callCount int
	var srvURL string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/page2" {
			_, _ = w.Write([]byte(page2))
		} else {
			_, _ = fmt.Fprintf(w, page1Tmpl, srvURL)
		}
	}))
	srvURL = srv.URL
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	watchers, err := client.ListRepoWatchers("ws", "repo")
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	require.Len(t, watchers, 2)
	assert.Equal(t, "alice", watchers[0].Slug)
	assert.Equal(t, "bob", watchers[1].Slug)
}
