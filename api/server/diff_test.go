package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

func newServerDiffClient(t *testing.T, handler http.HandlerFunc) *server.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL, "tok", "")
}

const serverDiffJSON = `{
  "diffs": [
    {
      "source":      {"toString":"api/foo.go"},
      "destination": {"toString":"api/foo.go"},
      "hunks": [
        {
          "sourceLine":1,"sourceSpan":2,"destinationLine":1,"destinationSpan":2,
          "segments": [
            {"type":"REMOVED","lines":[{"source":1,"destination":0,"line":"old line"}]},
            {"type":"ADDED",  "lines":[{"source":0,"destination":1,"line":"new line"}]},
            {"type":"CONTEXT","lines":[{"source":2,"destination":2,"line":"ctx line"}]}
          ]
        }
      ]
    }
  ]
}`

func TestServerClient_GetDiff_Path(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	client := newServerDiffClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"diffs":[]}`))
	})
	_, err := client.GetDiff("PROJ", "my-repo", "main", "feature")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/diff", gotPath)
	assert.Contains(t, gotQuery, "since=main")
	assert.Contains(t, gotQuery, "until=feature")
}

func TestServerClient_GetDiff_ReconstructsUnifiedDiff(t *testing.T) {
	t.Parallel()
	client := newServerDiffClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serverDiffJSON))
	})
	text, err := client.GetDiff("PROJ", "my-repo", "main", "feature")
	require.NoError(t, err)
	assert.Contains(t, text, "--- a/api/foo.go")
	assert.Contains(t, text, "+++ b/api/foo.go")
	assert.Contains(t, text, "-old line")
	assert.Contains(t, text, "+new line")
	assert.Contains(t, text, " ctx line")
}

func TestServerClient_GetDiffStat_Path(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	client := newServerDiffClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"diffs":[]}`))
	})
	_, err := client.GetDiffStat("PROJ", "my-repo", "main", "feature")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/diff", gotPath)
	assert.Contains(t, gotQuery, "since=main")
	assert.Contains(t, gotQuery, "until=feature")
}

func TestServerClient_GetDiffStat_CountsLinesAndFiles(t *testing.T) {
	t.Parallel()
	client := newServerDiffClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serverDiffJSON))
	})
	stat, err := client.GetDiffStat("PROJ", "my-repo", "main", "feature")
	require.NoError(t, err)
	assert.Equal(t, 1, stat.FilesChanged)
	assert.Equal(t, 1, stat.Additions)
	assert.Equal(t, 1, stat.Deletions)
	require.Len(t, stat.Files, 1)
	assert.Equal(t, "api/foo.go", stat.Files[0].Path)
	assert.Equal(t, "modified", stat.Files[0].Status)
}

func TestServerClient_GetDiffStat_StatusAdded(t *testing.T) {
	t.Parallel()
	const body = `{"diffs":[{"destination":{"toString":"new.go"},"hunks":[]}]}`
	client := newServerDiffClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
	stat, err := client.GetDiffStat("PROJ", "my-repo", "main", "feature")
	require.NoError(t, err)
	require.Len(t, stat.Files, 1)
	assert.Equal(t, "added", stat.Files[0].Status)
}

func TestServerClient_GetDiffStat_StatusDeleted(t *testing.T) {
	t.Parallel()
	const body = `{"diffs":[{"source":{"toString":"old.go"},"hunks":[]}]}`
	client := newServerDiffClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
	stat, err := client.GetDiffStat("PROJ", "my-repo", "main", "feature")
	require.NoError(t, err)
	require.Len(t, stat.Files, 1)
	assert.Equal(t, "deleted", stat.Files[0].Status)
}
