package httpx_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/internal/httpx"
)

// newTestTransport creates a Transport wired to handler using
// ContentTypeAlwaysWrite — the Bitbucket Server/DC default.
func newTestTransport(t *testing.T, handler http.HandlerFunc) (*httpx.Transport, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	tr := httpx.New(srv.Client(), srv.URL, httpx.Auth{Token: "tok"}, nil, httpx.ContentTypeAlwaysWrite, nil)
	return tr, srv
}

// ---------------------------------------------------------------------------
// ContentTypeAlwaysWrite policy (Bitbucket Server / CSRF protection)
// ---------------------------------------------------------------------------

// TestContentTypeAlwaysWrite_NilBodyPost_SetsContentType verifies that a
// nil-body POST still carries Content-Type so Bitbucket Server's CSRF filter
// does not reject the request with "XSRF check failed".
func TestContentTypeAlwaysWrite_NilBodyPost_SetsContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	tr, _ := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	var result struct{}
	require.NoError(t, tr.PostJSON("/test", nil, &result))
	assert.Equal(t, "application/json", gotCT,
		"ContentTypeAlwaysWrite: nil-body POST must set Content-Type for Bitbucket Server CSRF")
}

// TestContentTypeAlwaysWrite_NilBodyDelete_SetsContentType verifies the same
// for DELETE.
func TestContentTypeAlwaysWrite_NilBodyDelete_SetsContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	tr, _ := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusNoContent)
	})
	require.NoError(t, tr.DeleteJSON("/test", nil))
	assert.Equal(t, "application/json", gotCT,
		"ContentTypeAlwaysWrite: nil-body DELETE must set Content-Type for Bitbucket Server CSRF")
}

// ---------------------------------------------------------------------------
// ContentTypeWhenBody policy (Bitbucket Cloud)
// ---------------------------------------------------------------------------

// TestContentTypeWhenBody_NilBodyPost_NoContentType verifies that when the
// ContentTypeWhenBody policy is active a nil-body POST does NOT set
// Content-Type.  Bitbucket Cloud returns HTTP 400 when an empty POST includes
// Content-Type (e.g. ApprovePR, DeclinePR, RequestChangesPR).
func TestContentTypeWhenBody_NilBodyPost_NoContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	tr := httpx.New(srv.Client(), srv.URL, httpx.Auth{Token: "tok"}, nil, httpx.ContentTypeWhenBody, nil)
	var result struct{}
	require.NoError(t, tr.PostJSON("/test", nil, &result))
	assert.Empty(t, gotCT,
		"ContentTypeWhenBody: nil-body POST must NOT set Content-Type for Bitbucket Cloud")
}

// TestContentTypeWhenBody_WithBody_SetsContentType verifies that both policies
// correctly set Content-Type when a request body is present.
func TestContentTypeWhenBody_WithBody_SetsContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	tr := httpx.New(srv.Client(), srv.URL, httpx.Auth{Token: "tok"}, nil, httpx.ContentTypeWhenBody, nil)
	var result struct{}
	require.NoError(t, tr.PostJSON("/test", map[string]string{"k": "v"}, &result))
	assert.Equal(t, "application/json", gotCT,
		"ContentTypeWhenBody: body-present POST must set Content-Type")
}

// TestGetJSON_DoesNotSetContentType verifies that GET requests are unaffected
// regardless of policy.
func TestGetJSON_DoesNotSetContentType(t *testing.T) {
	t.Parallel()
	var gotCT string
	tr, _ := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	})
	var result struct{}
	require.NoError(t, tr.GetJSON("/test", &result))
	assert.Empty(t, gotCT, "GET must not set Content-Type")
}

// ---------------------------------------------------------------------------
// GetAllJSON pagination
// ---------------------------------------------------------------------------

// staticPaginator is a test Paginator that returns a predetermined sequence of
// next URLs, stopping (returning "") after the last one.
type staticPaginator struct {
	urls []string
	pos  int
}

func (p *staticPaginator) NextURL(_ string, _ []byte) string {
	if p.pos >= len(p.urls) {
		return ""
	}
	u := p.urls[p.pos]
	p.pos++
	return u
}

// TestGetAllJSON_FollowsPages verifies that GetAllJSON follows the Paginator's
// NextURL output across multiple pages and collects all bodies.
func TestGetAllJSON_FollowsPages(t *testing.T) {
	t.Parallel()
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"page":%d}`, callCount)
	}))
	t.Cleanup(srv.Close)

	pager := &staticPaginator{
		urls: []string{srv.URL + "/p2", srv.URL + "/p3"},
	}
	tr := httpx.New(srv.Client(), srv.URL, httpx.Auth{}, nil, httpx.ContentTypeWhenBody, pager)

	var pages []int
	err := tr.GetAllJSON("/p1", func(body []byte) error {
		var v struct {
			Page int `json:"page"`
		}
		if err := json.Unmarshal(body, &v); err != nil {
			return err
		}
		pages = append(pages, v.Page)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, pages)
	assert.Equal(t, 3, callCount)
}

// TestGetAllJSON_NilPaginator_FetchesOnce verifies that without a Paginator
// GetAllJSON fetches exactly one page and ignores any "next" field.
func TestGetAllJSON_NilPaginator_FetchesOnce(t *testing.T) {
	t.Parallel()
	callCount := 0
	tr, _ := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"next":"http://example.com/page2"}`))
	})
	err := tr.GetAllJSON("/items", func([]byte) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "without a Paginator only one request should be made")
}
