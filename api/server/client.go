// Package server is the Bitbucket Data Center (a.k.a. "Server") adapter for
// the backend.Client interface.
package server

import (
	"encoding/json"
	"io"
	"net/url"
	"strconv"
	"sync"

	"github.com/proggarapsody/bitbottle/api/backend"
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
	// defaultReviewersHTTP targets /rest/default-reviewers/1.0, the separate
	// REST root Bitbucket Server uses for matching the per-repo default-
	// reviewers conditions during PR creation.
	defaultReviewersHTTP *httpx.Transport
	// branchProtectHTTP targets /rest/branch-permissions/2.0, the separate
	// REST root Bitbucket Server uses for branch-restriction CRUD.
	branchProtectHTTP *httpx.Transport
	// codeInsightsHTTP targets /rest/insights/1.0, the separate REST root
	// Bitbucket Server uses for Code Insights reports and annotations.
	codeInsightsHTTP *httpx.Transport
	// host is the scheme+host extracted from baseURL, used to construct WebURLs
	// for resources (like commits) that the API does not return a link for.
	host string
	// userSlug is the authenticated user's slug. When non-empty it is used by
	// GetCurrentUser to call GET /users/{slug} instead of GET /users/~ because
	// Bitbucket Server does not recognise "~" as a self-reference.
	userSlug string
	// versionOnce guards the one-time fetch of the server version.
	versionOnce sync.Once
	// cachedVersion holds the parsed server version after the first successful fetch.
	cachedVersion backend.ServerVersion
}

// hostFromURL returns the bare hostname (no scheme) portion of a URL,
// or the URL itself if it cannot be parsed. Used to populate
// backend.DomainError.Host so the errfmt hint at pkg/errfmt/errfmt.go
// renders `--hostname HOST` rather than `--hostname https://HOST`
// (which would trip the scheme-stripping path in `auth login`).
// See PRD #372 Bug D.
func hostFromURL(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		return u.Host
	}
	return rawURL
}

// NewClient constructs a Client.
// If token is non-empty Bearer auth is used; else if username is non-empty
// Basic auth is used with username:token as credentials.
func NewClient(httpClient HTTPClient, baseURL, token, username string) *Client {
	// schemeHost is "scheme://host" used as the base for alt-transport
	// REST roots and for constructing WebURLs on resources (commits, etc.)
	// that the API does not return a self-link for.
	schemeHost := baseURL
	if u, err := url.Parse(baseURL); err == nil {
		schemeHost = u.Scheme + "://" + u.Host
	}
	// bareHost is the hostname only — used for DomainError.Host so the
	// errfmt hint renders a clean `--hostname HOST` (PRD #372 Bug D).
	bareHost := hostFromURL(baseURL)
	auth := httpx.Auth{Token: token, Username: username}
	return &Client{
		http: httpx.New(
			httpClient,
			baseURL,
			auth,
			decodeErrorMessage,
			httpx.ContentTypeAlwaysWrite,
			serverPaginator{},
		).UseDomainErrors(bareHost),
		buildStatusHTTP:      newAltTransport(httpClient, schemeHost, bareHost, "/rest/build-status/1.0", auth),
		defaultReviewersHTTP: newAltTransport(httpClient, schemeHost, bareHost, "/rest/default-reviewers/1.0", auth),
		branchProtectHTTP:    newAltTransport(httpClient, schemeHost, bareHost, "/rest/branch-permissions/2.0", auth),
		codeInsightsHTTP:     newAltTransport(httpClient, schemeHost, bareHost, "/rest/insights/1.0", auth),
		host:                 schemeHost,
		userSlug:             username,
	}
}

// newAltTransport constructs an httpx.Transport for a Bitbucket Server
// supplementary REST root (e.g. /rest/build-status/1.0). These roots share
// all transport config with the primary transport but target a different
// base URL. schemeHost is the full "scheme://host" baseURL prefix;
// bareHost is the hostname-only form used for DomainError.Host.
func newAltTransport(httpClient HTTPClient, schemeHost, bareHost, suffix string, auth httpx.Auth) *httpx.Transport {
	return httpx.New(
		httpClient,
		schemeHost+suffix,
		auth,
		decodeErrorMessage,
		httpx.ContentTypeAlwaysWrite,
		serverPaginator{},
	).UseDomainErrors(bareHost)
}

// PagedResponse is the Bitbucket Data Center paged list envelope.
type PagedResponse[T any] struct {
	Values        []T  `json:"values"`
	Size          int  `json:"size"`
	IsLastPage    bool `json:"isLastPage"`
	NextPageStart *int `json:"nextPageStart"`
	Start         int  `json:"start"`
}

// serverPaginator follows Bitbucket Server/DC pagination by inspecting
// isLastPage and nextPageStart, then appending start=N to the current URL.
type serverPaginator struct{}

func (serverPaginator) NextURL(currentURL string, responseBody []byte) string {
	var page struct {
		IsLastPage    bool `json:"isLastPage"`
		NextPageStart *int `json:"nextPageStart"`
	}
	if json.Unmarshal(responseBody, &page) != nil || page.IsLastPage || page.NextPageStart == nil {
		return ""
	}
	u, err := url.Parse(currentURL)
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Set("start", strconv.Itoa(*page.NextPageStart))
	u.RawQuery = q.Encode()
	return u.String()
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

func (c *Client) getJSON(path string, v any) error {
	return c.http.GetJSON(path, v)
}

func (c *Client) getText(path string) (string, error) {
	return c.http.GetText(path)
}

func (c *Client) postJSON(path string, body, v any) error {
	return c.http.PostJSON(path, body, v)
}

func (c *Client) putJSON(path string, body, v any) error {
	return c.http.PutJSON(path, body, v)
}

func (c *Client) delete(path string, body any) error {
	return c.http.DeleteJSON(path, body)
}
