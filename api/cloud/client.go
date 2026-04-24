// Package cloud is the Bitbucket Cloud adapter for the backend.Client
// interface.
package cloud

import (
	"encoding/json"
	"io"

	"github.com/proggarapsody/bitbottle/api/internal/httpx"
)

// HTTPClient is the transport interface for making HTTP requests.
type HTTPClient = httpx.Doer

// Client is the Bitbucket Cloud HTTP client.
type Client struct {
	http *httpx.Transport
}

// NewClient constructs a Cloud Client.
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

// cloudPagedResponse is the Bitbucket Cloud paged list envelope.
type cloudPagedResponse[T any] struct {
	Values []T    `json:"values"`
	Next   string `json:"next"`
}

// cloudErrorEnvelope is the Bitbucket Cloud error body shape.
type cloudErrorEnvelope struct {
	Type  string `json:"type"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// decodeErrorMessage parses a Cloud error response body and returns the
// message, or empty string if the body cannot be decoded.
func decodeErrorMessage(body io.Reader) string {
	var env cloudErrorEnvelope
	_ = json.NewDecoder(body).Decode(&env)
	return env.Error.Message
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

// delete sends a DELETE request to path (no body).
func (c *Client) delete(path string) error {
	return c.http.DeleteJSON(path, nil)
}
