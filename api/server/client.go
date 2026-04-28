// Package server is the Bitbucket Data Center (a.k.a. "Server") adapter for
// the backend.Client interface.
package server

import (
	"encoding/json"
	"io"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/internal/httpx"
)

// HTTPClient is the transport interface for making HTTP requests. It is
// retained as an alias at the package boundary so callers in this repository
// continue to compile without importing the internal httpx package.
type HTTPClient = httpx.Doer

// Client is the Bitbucket Data Center HTTP client.
type Client struct {
	http *httpx.Transport
	// buildStatusHTTP targets /rest/build-status/1.0, the separate REST root
	// Bitbucket Server uses for commit build statuses.
	buildStatusHTTP *httpx.Transport
	// host is the scheme+host extracted from baseURL, used to construct WebURLs
	// for resources (like commits) that the API does not return a link for.
	host string
	// userSlug is the authenticated user's slug. When non-empty it is used by
	// GetCurrentUser to call GET /users/{slug} instead of GET /users/~ because
	// Bitbucket Server does not recognise "~" as a self-reference.
	userSlug string
}

// NewClient constructs a Client.
// If token is non-empty Bearer auth is used; else if username is non-empty
// Basic auth is used with username:token as credentials.
func NewClient(httpClient HTTPClient, baseURL, token, username string) *Client {
	host := baseURL
	if u, err := url.Parse(baseURL); err == nil {
		host = u.Scheme + "://" + u.Host
	}
	auth := httpx.Auth{Token: token, Username: username}
	return &Client{
		http: httpx.New(
			httpClient,
			baseURL,
			auth,
			decodeErrorMessage,
		),
		buildStatusHTTP: httpx.New(
			httpClient,
			host+"/rest/build-status/1.0",
			auth,
			decodeErrorMessage,
		),
		host:     host,
		userSlug: username,
	}
}

// PagedResponse is the Bitbucket Data Center paged list envelope.
type PagedResponse[T any] struct {
	Values        []T  `json:"values"`
	Size          int  `json:"size"`
	IsLastPage    bool `json:"isLastPage"`
	NextPageStart *int `json:"nextPageStart"`
	Start         int  `json:"start"`
}

// dcErrorEnvelope is the Bitbucket Data Center error body shape.
type dcErrorEnvelope struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// decodeErrorMessage parses a Data Center error response body and returns the
// first error message, or empty string if the body cannot be decoded.
func decodeErrorMessage(body io.Reader) string {
	var env dcErrorEnvelope
	_ = json.NewDecoder(body).Decode(&env)
	if len(env.Errors) == 0 {
		return ""
	}
	return env.Errors[0].Message
}

// getJSON GETs path and decodes the JSON response into v.
func (c *Client) getJSON(path string, v any) error {
	return c.http.GetJSON(path, v)
}

// getText GETs path and returns the raw body string.
func (c *Client) getText(path string) (string, error) {
	return c.http.GetText(path)
}

// postJSON POSTs body as JSON to path and decodes the response into v.
func (c *Client) postJSON(path string, body, v any) error {
	return c.http.PostJSON(path, body, v)
}

// putJSON PUTs body as JSON to path and decodes the response into v.
func (c *Client) putJSON(path string, body, v any) error {
	return c.http.PutJSON(path, body, v)
}

// delete sends a DELETE request to path with an optional JSON body.
func (c *Client) delete(path string, body any) error {
	return c.http.DeleteJSON(path, body)
}
