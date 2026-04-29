// Package cloud is the Bitbucket Cloud adapter for the backend.Client
// interface.
package cloud

import (
	"encoding/json"
	"io"
	"net/url"
	"regexp"

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
			httpx.ContentTypeWhenBody,
			cloudPaginator{},
		),
	}
}

// cloudPagedResponse is the Bitbucket Cloud paged list envelope.
type cloudPagedResponse[T any] struct {
	Values []T    `json:"values"`
	Next   string `json:"next"`
}

// cloudPaginator follows Bitbucket Cloud pagination by extracting the "next"
// absolute URL from the paged response envelope.
type cloudPaginator struct{}

func (cloudPaginator) NextURL(_ string, responseBody []byte) string {
	var page struct {
		Next string `json:"next"`
	}
	if json.Unmarshal(responseBody, &page) != nil || page.Next == "" {
		return ""
	}
	// Validate that next is a well-formed absolute HTTP(S) URL; malformed
	// values (e.g. "::::not-a-url::::") must not cause a panic or an error.
	u, err := url.ParseRequestURI(page.Next)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	return page.Next
}

// cloudErrorEnvelope is the Bitbucket Cloud error body shape.
type cloudErrorEnvelope struct {
	Type  string `json:"type"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// fieldPrefixRE matches a leading "fieldname: " prefix that Bitbucket Cloud
// sometimes prepends to error messages, e.g. "newstatus: Already closed."
var fieldPrefixRE = regexp.MustCompile(`^[a-z_]+:\s+`)

// decodeErrorMessage parses a Cloud error response body and returns the
// message, or empty string if the body cannot be decoded.
// Bitbucket Cloud occasionally prefixes the human-readable message with the
// internal field name (e.g. "newstatus: This pull request is already closed.");
// that prefix is stripped so callers see only the human-readable portion.
func decodeErrorMessage(body io.Reader) string {
	var env cloudErrorEnvelope
	_ = json.NewDecoder(body).Decode(&env)
	msg := env.Error.Message
	return fieldPrefixRE.ReplaceAllString(msg, "")
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

// delete sends a DELETE request to path (no body).
func (c *Client) delete(path string) error {
	return c.http.DeleteJSON(path, nil)
}
