// Package server is the Bitbucket Data Center (a.k.a. "Server") adapter for
// the backend.Client interface.
package server

import (
	"encoding/json"
	"io"

	"github.com/proggarapsody/bitbottle/api/internal/httpx"
)

// HTTPClient is the transport interface for making HTTP requests. It is
// retained as an alias at the package boundary so callers in this repository
// continue to compile without importing the internal httpx package.
type HTTPClient = httpx.Doer

// Client is the Bitbucket Data Center HTTP client.
type Client struct {
	http *httpx.Transport
}

// NewClient constructs a Client.
// If token is non-empty Bearer auth is used; else if username is non-empty
// Basic auth is used with username:token as credentials.
func NewClient(httpClient HTTPClient, baseURL, token, username string) *Client {
	return &Client{
		http: httpx.New(
			httpClient,
			baseURL,
			httpx.Auth{Token: token, Username: username},
			decodeErrorMessage,
		),
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
