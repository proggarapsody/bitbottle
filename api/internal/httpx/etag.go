package httpx

import (
	"bytes"
	"container/list"
	"io"
	"net/http"
	"strings"
	"sync"
)

const defaultETagBudget = 16 * 1024 * 1024 // 16 MB

// eTagEntry is one cached response.
type eTagEntry struct {
	url  string
	etag string
	body []byte
	size int
}

// eTagCache is a byte-budget LRU. Evicts oldest entries when budget exceeded.
type eTagCache struct {
	mu      sync.Mutex
	budget  int
	used    int
	entries map[string]*list.Element
	order   *list.List
}

func newETagCache(budget int) *eTagCache {
	return &eTagCache{
		budget:  budget,
		entries: map[string]*list.Element{},
		order:   list.New(),
	}
}

// get returns cached entry or nil. Marks it most-recently-used.
func (c *eTagCache) get(url string) *eTagEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.entries[url]
	if !ok {
		return nil
	}
	c.order.MoveToFront(el)
	entry := el.Value.(*eTagEntry)
	// return copy of body to avoid races on slice contents
	cp := *entry
	return &cp
}

// set stores entry. Evicts LRU entries if budget exceeded.
func (c *eTagCache) set(url string, entry eTagEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok := c.entries[url]; ok {
		old := existing.Value.(*eTagEntry)
		c.used -= old.size
		c.order.Remove(existing)
		delete(c.entries, url)
	}
	entry.url = url
	el := c.order.PushFront(&entry)
	c.entries[url] = el
	c.used += entry.size
	for c.used > c.budget && c.order.Len() > 0 {
		back := c.order.Back()
		if back == nil {
			break
		}
		oldEntry := back.Value.(*eTagEntry)
		c.used -= oldEntry.size
		c.order.Remove(back)
		delete(c.entries, oldEntry.url)
	}
}

// invalidatePrefix removes all entries whose URL key starts with prefix.
func (c *eTagCache) invalidatePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, el := range c.entries {
		if strings.HasPrefix(k, prefix) {
			entry := el.Value.(*eTagEntry)
			c.used -= entry.size
			c.order.Remove(el)
			delete(c.entries, k)
		}
	}
}

// eTagRoundTripper wraps inner and adds ETag caching for GET requests, plus
// prefix-based invalidation on non-GET 2xx responses.
type eTagRoundTripper struct {
	inner http.RoundTripper
	cache *eTagCache
}

func (r *eTagRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.URL.String()
	if req.Method == http.MethodGet {
		if cached := r.cache.get(key); cached != nil {
			// Clone before mutating: RoundTripper contract forbids modifying the
			// caller's request. Clone preserves Body and GetBody references.
			reqCopy := req.Clone(req.Context())
			reqCopy.Header.Set("If-None-Match", cached.etag)
			resp, err := r.inner.RoundTrip(reqCopy)
			if err != nil {
				return nil, err
			}
			if resp.StatusCode == http.StatusNotModified {
				_ = resp.Body.Close()
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     cloneHeader(resp.Header),
					Body:       io.NopCloser(bytes.NewReader(cached.body)),
					Request:    req,
				}, nil
			}
			return r.cacheIfETag(resp, key)
		}
		resp, err := r.inner.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		return r.cacheIfETag(resp, key)
	}
	// non-GET: pass through, then invalidate on success
	resp, err := r.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// invalidate the exact URL and any sub-paths
		r.cache.invalidatePrefix(key)
	}
	return resp, nil
}

// cacheIfETag stores the response body in cache if the response is 2xx and has
// an ETag header. It always returns a response whose Body is fully readable.
func (r *eTagRoundTripper) cacheIfETag(resp *http.Response, key string) (*http.Response, error) {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, nil
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		return resp, nil
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	r.cache.set(key, eTagEntry{
		etag: etag,
		body: body,
		size: len(body),
	})
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}

func cloneHeader(h http.Header) http.Header {
	out := http.Header{}
	for k, v := range h {
		out[k] = append([]string{}, v...)
	}
	return out
}
