// Package httpx is the shared HTTP transport used by the Bitbucket Server/DC
// and Cloud API adapters. It handles request construction, authentication
// injection, JSON encoding/decoding, and error translation into
// backend.HTTPError.
//
// Each adapter plugs in its own ErrorDecoder to parse backend-specific error
// response bodies, a ContentTypePolicy to control when Content-Type is set,
// and a Paginator to follow multi-page results.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// Doer is the transport interface for making HTTP requests. It is satisfied by
// *http.Client.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Auth holds credentials for a single host. If Token is non-empty Bearer auth
// is used; otherwise if Username is non-empty Basic auth is used with
// Username:Token as credentials.
type Auth struct {
	Token    string
	Username string
}

// ErrorDecoder extracts a human-readable message from an error response body.
// Different Bitbucket products return different error envelopes, so each
// adapter supplies its own decoder.
type ErrorDecoder func(body io.Reader) string

// ContentTypePolicy controls when the Content-Type: application/json header is
// added to a request. method is the HTTP method; hasBody is true when the
// request carries a non-nil body.
//
// Use ContentTypeWhenBody for Bitbucket Cloud and ContentTypeAlwaysWrite for
// Bitbucket Server/DC.
type ContentTypePolicy func(method string, hasBody bool) bool

// ContentTypeWhenBody sets Content-Type only when the request carries a body.
// Use this for Bitbucket Cloud, which returns HTTP 400 when an empty-body
// POST/PUT includes a Content-Type header (e.g. ApprovePR, DeclinePR,
// RequestChangesPR).
var ContentTypeWhenBody ContentTypePolicy = func(_ string, hasBody bool) bool {
	return hasBody
}

// ContentTypeAlwaysWrite sets Content-Type for every write method (POST, PUT,
// DELETE) even when the body is nil. Use this for Bitbucket Server/Data Center
// whose CSRF filter rejects write requests that omit Content-Type.
var ContentTypeAlwaysWrite ContentTypePolicy = func(method string, hasBody bool) bool {
	return hasBody || (method != http.MethodGet && method != http.MethodHead)
}

// Paginator extracts the next-page URL from a response body. Implementations
// live in each adapter package because the pagination envelope differs between
// Cloud ("next": "<absolute-url>") and Server/DC ("isLastPage": bool,
// "nextPageStart": N).
type Paginator interface {
	// NextURL returns the absolute URL of the next page, or "" when there are
	// no more pages. currentURL is the absolute URL that produced responseBody.
	NextURL(currentURL string, responseBody []byte) string
}

// Transport encapsulates auth injection and JSON helpers over a Doer.
type Transport struct {
	doer              Doer
	baseURL           string
	auth              Auth
	decodeErrMsg      ErrorDecoder
	contentTypePolicy ContentTypePolicy
	paginator         Paginator
	domainHost        string // when non-empty, apiError wraps via backend.ClassifyHTTPError
}

// UseDomainErrors enables classification of HTTP errors into typed
// backend.DomainError values on the way out of GetJSON / PostJSON / pagination
// helpers. host is attached to the returned DomainError so callers and the
// MCP surface know which Bitbucket instance produced the failure.
//
// Adapters call this once during construction. When unset (the default),
// apiError continues to return the raw *backend.HTTPError for back-compat
// with existing tests.
func (t *Transport) UseDomainErrors(host string) *Transport {
	t.domainHost = host
	return t
}

// New constructs a Transport.
//
// ctPolicy controls Content-Type header injection; pass ContentTypeWhenBody
// for Bitbucket Cloud or ContentTypeAlwaysWrite for Bitbucket Server/DC.
//
// paginator is used by GetAllJSON to follow pages; pass nil if pagination is
// not required for this transport instance.
func New(doer Doer, baseURL string, auth Auth, decodeErr ErrorDecoder, ctPolicy ContentTypePolicy, paginator Paginator) *Transport {
	return &Transport{
		doer:              doer,
		baseURL:           strings.TrimRight(baseURL, "/"),
		auth:              auth,
		decodeErrMsg:      decodeErr,
		contentTypePolicy: ctPolicy,
		paginator:         paginator,
	}
}

// do adds auth headers and executes the request.
//
// Auth priority:
//  1. Both Username and Token set → Basic auth (username:token).
//     This covers Bitbucket App Passwords and Atlassian API tokens, which both
//     require HTTP Basic auth rather than Bearer.
//  2. Token only → Bearer auth.
//     Used for OAuth2 access tokens and workspace/repository access tokens.
//  3. Username only → Basic auth with empty password (rare, kept for compat).
func (t *Transport) do(req *http.Request) (*http.Response, error) {
	switch {
	case t.auth.Username != "" && t.auth.Token != "":
		req.SetBasicAuth(t.auth.Username, t.auth.Token)
	case t.auth.Token != "":
		req.Header.Set("Authorization", "Bearer "+t.auth.Token)
	case t.auth.Username != "":
		req.SetBasicAuth(t.auth.Username, t.auth.Token)
	}
	return t.doer.Do(req)
}

func (t *Transport) GetJSON(path string, v any) error {
	return t.sendJSON(http.MethodGet, path, nil, v)
}

// GetAllJSON fetches all pages starting at path by following the Transport's
// Paginator (if any). accumulate is called with the raw JSON bytes for each
// page. If no Paginator is configured only the first page is fetched.
//
// The first request uses t.baseURL+path. Subsequent requests use the absolute
// URL returned by Paginator.NextURL, which lets each adapter embed the correct
// host in its next-page field.
func (t *Transport) GetAllJSON(path string, accumulate func([]byte) error) error {
	nextURL := t.baseURL + path
	for nextURL != "" {
		body, err := t.fetchBodyAt(nextURL)
		if err != nil {
			return err
		}
		if err := accumulate(body); err != nil {
			return err
		}
		if t.paginator == nil {
			break
		}
		nextURL = t.paginator.NextURL(nextURL, body)
	}
	return nil
}

func (t *Transport) fetchBodyAt(fullURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, t.apiError(resp)
	}
	return io.ReadAll(resp.Body)
}

// GetText GETs path and returns the raw body string.
// Sends Accept: text/plain so Bitbucket Server returns a unified diff
// rather than its JSON structured-diff format.
func (t *Transport) GetText(path string) (string, error) {
	req, err := t.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")
	resp, err := t.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode >= http.StatusBadRequest {
		return "", t.apiError(resp)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

func (t *Transport) PostJSON(path string, body, v any) error {
	return t.sendJSON(http.MethodPost, path, body, v)
}

func (t *Transport) PutJSON(path string, body, v any) error {
	return t.sendJSON(http.MethodPut, path, body, v)
}

// DeleteJSON sends a DELETE request with an optional JSON body (body may be nil).
func (t *Transport) DeleteJSON(path string, body any) error {
	return t.sendJSON(http.MethodDelete, path, body, nil)
}

func (t *Transport) sendJSON(method, path string, body, v any) error {
	req, err := t.newRequest(method, path, body)
	if err != nil {
		return err
	}
	resp, err := t.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	return t.checkAndDecode(resp, v)
}

func (t *Transport) newRequest(method, path string, body any) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, t.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	if t.contentTypePolicy != nil && t.contentTypePolicy(method, body != nil) {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (t *Transport) checkAndDecode(resp *http.Response, v any) error {
	if resp.StatusCode >= http.StatusBadRequest {
		return t.apiError(resp)
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func (t *Transport) apiError(resp *http.Response) error {
	msg := ""
	if t.decodeErrMsg != nil {
		msg = t.decodeErrMsg(resp.Body)
	}
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}
	url := ""
	if resp.Request != nil {
		url = resp.Request.URL.String()
	}
	httpErr := &backend.HTTPError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		RequestURL: url,
	}
	if t.domainHost == "" {
		return httpErr
	}
	return backend.ClassifyHTTPError(t.domainHost, httpErr)
}
