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

const listRepoForksJSON = `{"values":[{"full_name":"forkws/my-service","name":"my-service","slug":"my-service","scm":"git","is_private":true,"description":"A fork","links":{"html":{"href":"https://bitbucket.org/forkws/my-service"}}}]}`

func TestCloudClient_ListRepoForks_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.ListRepoForks("myworkspace", "my-service", 10)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/forks", gotPath)
}

func TestCloudClient_ListRepoForks_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listRepoForksJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	forks, err := client.ListRepoForks("myworkspace", "my-service", 10)
	require.NoError(t, err)
	require.Len(t, forks, 1)
	assert.Equal(t, "my-service", forks[0].Slug)
	assert.Equal(t, "my-service", forks[0].Name)
	assert.Equal(t, "forkws", forks[0].Namespace)
	assert.Equal(t, "git", forks[0].SCM)
	assert.True(t, forks[0].IsPrivate)
	assert.Equal(t, "A fork", forks[0].Description)
	assert.Contains(t, forks[0].WebURL, "forkws/my-service")
}

func TestCloudClient_ListRepoForks_FollowsPagination(t *testing.T) {
	t.Parallel()
	page1Tmpl := `{"values":[{"full_name":"fws/fork-a","name":"fork-a","slug":"fork-a","scm":"git","links":{"html":{"href":""}}}],"next":"%s/page2"}`
	page2 := `{"values":[{"full_name":"fws/fork-b","name":"fork-b","slug":"fork-b","scm":"git","links":{"html":{"href":""}}}]}`

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

	forks, err := client.ListRepoForks("ws", "repo", 0)
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	require.Len(t, forks, 2)
	assert.Equal(t, "fork-a", forks[0].Slug)
	assert.Equal(t, "fork-b", forks[1].Slug)
}
