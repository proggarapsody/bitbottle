// Package httpx is the shared HTTP transport used by the Bitbucket Server/DC
// and Cloud API adapters. It handles request construction, authentication
// injection, JSON encoding/decoding, and error translation into
// backend.HTTPError.
//
// Each adapter plugs in its own ErrorDecoder to parse backend-specific error
// response bodies.
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

// Transport encapsulates auth injection and JSON helpers over a Doer.
type Transport struct {
	doer         Doer
	baseURL      string
	auth         Auth
	decodeErrMsg ErrorDecoder
}

// New constructs a Transport.
func New(doer Doer, baseURL string, auth Auth, decodeErr ErrorDecoder) *Transport {
	return &Transport{
		doer:         doer,
		baseURL:      strings.TrimRight(baseURL, "/"),
		auth:         auth,
		decodeErrMsg: decodeErr,
	}
}

// do adds auth headers and executes the request.
func (t *Transport) do(req *http.Request) (*http.Response, error) {
	switch {
	case t.auth.Token != "":
		req.Header.Set("Authorization", "Bearer "+t.auth.Token)
	case t.auth.Username != "":
		req.SetBasicAuth(t.auth.Username, t.auth.Token)
	}
	return t.doer.Do(req)
}

// GetJSON GETs path and decodes the JSON response into v.
func (t *Transport) GetJSON(path string, v any) error {
	return t.sendJSON(http.MethodGet, path, nil, v)
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

// PostJSON POSTs body as JSON to path and decodes the response into v.
func (t *Transport) PostJSON(path string, body, v any) error {
	return t.sendJSON(http.MethodPost, path, body, v)
}

// PutJSON PUTs body as JSON to path and decodes the response into v.
func (t *Transport) PutJSON(path string, body, v any) error {
	return t.sendJSON(http.MethodPut, path, body, v)
}

// DeleteJSON sends a DELETE request with an optional JSON body (body may be
// nil). It decodes nothing from the response.
func (t *Transport) DeleteJSON(path string, body any) error {
	return t.sendJSON(http.MethodDelete, path, body, nil)
}

// sendJSON handles every method that may carry a JSON body. It marshals body
// (if non-nil), sets Content-Type, sends the request and then checks/decodes
// the response into v (v may be nil).
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

// newRequest builds an *http.Request with the JSON body (if any) encoded and
// Content-Type set appropriately.
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
	// Always set Content-Type for write methods so that Bitbucket Server's
	// CSRF protection (which triggers on POST/PUT/DELETE without the header)
	// does not reject requests that carry no body (e.g. ApprovePR, DeclinePR).
	if body != nil || (method != http.MethodGet && method != http.MethodHead) {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// checkAndDecode returns an error for non-2xx responses, otherwise decodes
// into v (unless v is nil).
func (t *Transport) checkAndDecode(resp *http.Response, v any) error {
	if resp.StatusCode >= http.StatusBadRequest {
		return t.apiError(resp)
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// apiError builds a backend.HTTPError from a non-2xx response, using the
// adapter's ErrorDecoder to extract the backend-specific message.
func (t *Transport) apiError(resp *http.Response) error {
	msg := ""
	if t.decodeErrMsg != nil {
		msg = t.decodeErrMsg(resp.Body)
	}
	url := ""
	if resp.Request != nil {
		url = resp.Request.URL.String()
	}
	return &backend.HTTPError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		RequestURL: url,
	}
}
