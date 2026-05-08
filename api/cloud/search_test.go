package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

// codeSearchPage1 is the canonical wire shape Cloud returns for a code
// search hit: path_matches + content_matches with nested lines, plus the
// file/commit/repository nesting whose full_name we use as the domain
// "Repository" string. The "next" link drives multi-page collection.
const codeSearchPage1 = `{
  "values": [
    {
      "type": "code_search_result",
      "content_match_count": 2,
      "path_matches": [
        {"text": "src/"},
        {"text": "READ", "match": true},
        {"text": "ME.md"}
      ],
      "content_matches": [
        {
          "lines": [
            {
              "line": 7,
              "segments": [
                {"text": "func "},
                {"text": "TODO", "match": true},
                {"text": "Foo() {"}
              ]
            },
            {
              "line": 8,
              "segments": [
                {"text": "  // "},
                {"text": "TODO", "match": true},
                {"text": ": refactor"}
              ]
            }
          ]
        }
      ],
      "file": {
        "path": "src/README.md",
        "commit": {
          "repository": {
            "full_name": "acme/widgets"
          }
        },
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/acme/widgets/src/HEAD/src/README.md"
          }
        }
      }
    }
  ],
  "next": "%s/workspaces/acme/search/code?search_query=TODO&page=2"
}`

const codeSearchPage2 = `{
  "values": [
    {
      "type": "code_search_result",
      "content_match_count": 0,
      "path_matches": [
        {"text": "scripts/build.sh", "match": true}
      ],
      "content_matches": [],
      "file": {
        "path": "scripts/build.sh",
        "commit": {
          "repository": {
            "full_name": "acme/widgets"
          }
        },
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/acme/widgets/src/HEAD/scripts/build.sh"
          }
        }
      }
    }
  ]
}`

func newCloudSearchServer(t *testing.T, handler http.HandlerFunc) (*cloud.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", ""), srv
}

func TestCloudClient_SearchCode_PathAndQuery(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	client, _ := newCloudSearchServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	_, err := client.SearchCode("acme", "TODO", 25)
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/acme/search/code", gotPath)
	assert.Contains(t, gotQuery, "search_query=TODO")
	assert.Contains(t, gotQuery, "pagelen=25")
}

func TestCloudClient_SearchCode_RequiresWorkspaceAndQuery(t *testing.T) {
	t.Parallel()
	client := cloud.NewClient(http.DefaultClient, "https://example.invalid", "tok", "")
	_, err := client.SearchCode("", "x", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace")

	_, err = client.SearchCode("acme", "", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "query")
}

func TestCloudClient_SearchCode_DecodesAndFlattensLines(t *testing.T) {
	t.Parallel()
	client, srv := newCloudSearchServer(t, nil)
	// Two-page exchange: page 1 returns next pointing at page 2.
	mux := http.NewServeMux()
	mux.HandleFunc("/workspaces/acme/search/code", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(codeSearchPage2))
			return
		}
		// Substitute %s with the test server URL so the absolute "next"
		// link points back to ourselves.
		body := []byte(replace(codeSearchPage1, "%s", srv.URL))
		_, _ = w.Write(body)
	})
	srv.Config.Handler = mux

	got, err := client.SearchCode("acme", "TODO", 0)
	require.NoError(t, err)
	require.Len(t, got, 2, "two pages must flatten into two hits")

	// Hit 0: README with two matched lines on a single content_matches group.
	hit := got[0]
	assert.Equal(t, "acme/widgets", hit.Repository)
	assert.Equal(t, "src/README.md", hit.Path)
	assert.Equal(t, 2, hit.ContentMatchCount)
	require.Len(t, hit.PathMatches, 3)
	assert.Equal(t, "READ", hit.PathMatches[1].Text)
	assert.True(t, hit.PathMatches[1].Match, "matched segment must keep its match flag")

	require.Len(t, hit.ContentMatches, 2, "lines must flatten across the lines[] array")
	assert.Equal(t, 7, hit.ContentMatches[0].Line)
	assert.Equal(t, 8, hit.ContentMatches[1].Line)
	require.Len(t, hit.ContentMatches[0].Segments, 3)
	assert.True(t, hit.ContentMatches[0].Segments[1].Match)
	assert.Equal(t, "TODO", hit.ContentMatches[0].Segments[1].Text)
	assert.NotEmpty(t, hit.FileURL)

	// Hit 1: path-only match, no content_matches lines.
	assert.Equal(t, "scripts/build.sh", got[1].Path)
	assert.Empty(t, got[1].ContentMatches)
}

func TestCloudClient_SearchCode_LimitClampsTotalItems(t *testing.T) {
	t.Parallel()
	client, srv := newCloudSearchServer(t, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/workspaces/acme/search/code", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(codeSearchPage2))
			return
		}
		_, _ = w.Write([]byte(replace(codeSearchPage1, "%s", srv.URL)))
	})
	srv.Config.Handler = mux

	// Two pages produce two hits in total; limit=1 must clamp the slice.
	got, err := client.SearchCode("acme", "TODO", 1)
	require.NoError(t, err)
	require.Len(t, got, 1, "limit must cap total items even when more pages exist")
	assert.Equal(t, "src/README.md", got[0].Path)
}

// replace substitutes the test-server URL into the canned "next" link.
// Aliased to strings.ReplaceAll so the fixture stays declarative.
func replace(s, old, new string) string { return strings.ReplaceAll(s, old, new) }

// A workspace slug with a literal "/" must be percent-encoded into the URL
// path so a hostile or buggy caller can't pivot to a different endpoint.
// Mirrors the url.PathEscape pattern used in tag.go / prs.go.
func TestCloudClient_SearchCode_WorkspaceIsPathEscaped(t *testing.T) {
	t.Parallel()
	var gotEscapedPath string
	client, _ := newCloudSearchServer(t, func(w http.ResponseWriter, r *http.Request) {
		// EscapedPath preserves %2F; r.URL.Path would silently decode it.
		gotEscapedPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	_, err := client.SearchCode("acme/path", "TODO", 0)
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/acme%2Fpath/search/code", gotEscapedPath,
		"workspace must be url.PathEscape'd so '/' does not split the path")
}
