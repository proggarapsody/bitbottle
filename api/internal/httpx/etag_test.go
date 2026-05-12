package httpx

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestETagCache_SetGet(t *testing.T) {
	c := newETagCache(1024)
	c.set("/x", eTagEntry{etag: `"abc"`, body: []byte("hello"), size: 5})
	got := c.get("/x")
	require.NotNil(t, got)
	assert.Equal(t, `"abc"`, got.etag)
	assert.Equal(t, []byte("hello"), got.body)
}

func TestETagCache_ByteBudget_Evicts(t *testing.T) {
	c := newETagCache(20) // budget 20 bytes
	c.set("/a", eTagEntry{etag: `"a"`, body: []byte("0123456789"), size: 10})
	c.set("/b", eTagEntry{etag: `"b"`, body: []byte("0123456789"), size: 10})
	// both fit; budget exhausted but not exceeded
	require.NotNil(t, c.get("/a"))
	require.NotNil(t, c.get("/b"))
	// adding /c forces eviction of LRU. After get(/b) above, /b is most recent.
	// LRU at this point is /a.
	c.set("/c", eTagEntry{etag: `"c"`, body: []byte("0123456789"), size: 10})
	assert.Nil(t, c.get("/a"))
	assert.NotNil(t, c.get("/b"))
	assert.NotNil(t, c.get("/c"))
}

func TestETagCache_InvalidatePrefix(t *testing.T) {
	c := newETagCache(1024)
	c.set("/a/b", eTagEntry{etag: `"1"`, body: []byte("x"), size: 1})
	c.set("/a/c", eTagEntry{etag: `"2"`, body: []byte("x"), size: 1})
	c.set("/other", eTagEntry{etag: `"3"`, body: []byte("x"), size: 1})
	c.invalidatePrefix("/a")
	assert.Nil(t, c.get("/a/b"))
	assert.Nil(t, c.get("/a/c"))
	assert.NotNil(t, c.get("/other"))
}

func TestETagRoundTripper_Miss_CachesResponse(t *testing.T) {
	inner := &fakeRoundTripper{
		responses: []*http.Response{{
			StatusCode: 200,
			Header:     http.Header{"Etag": []string{`"v1"`}},
			Body:       io.NopCloser(strings.NewReader("payload")),
		}},
	}
	cache := newETagCache(1024)
	rt := &eTagRoundTripper{inner: inner, cache: cache}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/x", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, "payload", string(body))
	// cached now
	entry := cache.get("http://example.com/x")
	require.NotNil(t, entry)
	assert.Equal(t, `"v1"`, entry.etag)
	assert.Equal(t, []byte("payload"), entry.body)
}

func TestETagRoundTripper_Hit_Sends304(t *testing.T) {
	cache := newETagCache(1024)
	cache.set("http://example.com/x", eTagEntry{
		etag: `"v1"`,
		body: []byte("cached-body"),
		size: 11,
	})
	inner := &fakeRoundTripper{
		responses: []*http.Response{{
			StatusCode: 304,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("")),
		}},
	}
	rt := &eTagRoundTripper{inner: inner, cache: cache}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/x", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, `"v1"`, inner.lastReq.Header.Get("If-None-Match"))
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, "cached-body", string(body))
}

func TestETagRoundTripper_WriteInvalidates(t *testing.T) {
	cache := newETagCache(1024)
	cache.set("http://example.com/repos/foo", eTagEntry{
		etag: `"v1"`,
		body: []byte("x"),
		size: 1,
	})
	inner := &fakeRoundTripper{
		responses: []*http.Response{{
			StatusCode: 201,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader("")),
		}},
	}
	rt := &eTagRoundTripper{inner: inner, cache: cache}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://example.com/repos/foo", strings.NewReader(""))
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Nil(t, cache.get("http://example.com/repos/foo"))
}
