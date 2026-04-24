package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPClient is the interface for making HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// AuthConfig holds credentials for one host.
type AuthConfig struct {
	Token    string
	Username string
	Password string
}

// Client wraps HTTPClient with auth injection and JSON helpers.
type Client struct {
	http    HTTPClient
	baseURL string
	auth    AuthConfig
}

// NewClient constructs a Client.
func NewClient(httpClient HTTPClient, baseURL string, auth AuthConfig) *Client {
	return &Client{http: httpClient, baseURL: strings.TrimRight(baseURL, "/"), auth: auth}
}

// HTTPError represents a Bitbucket API error response.
type HTTPError struct {
	StatusCode int
	Message    string
	RequestURL string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	if c.auth.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.auth.Token)
	} else if c.auth.Username != "" {
		req.SetBasicAuth(c.auth.Username, c.auth.Password)
	}
	return c.http.Do(req)
}

// GetJSON GETs path and decodes JSON into v.
func (c *Client) GetJSON(path string, v any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.checkAndDecode(resp, v)
}

// GetText GETs path and returns the raw body as a string.
func (c *Client) GetText(path string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", c.apiError(resp)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// PostJSON POSTs body as JSON to path and decodes response into v.
func (c *Client) PostJSON(path string, body any, v any) error {
	return c.doJSON(http.MethodPost, path, body, v)
}

// PutJSON PUTs body as JSON to path and decodes response into v.
func (c *Client) PutJSON(path string, body any, v any) error {
	return c.doJSON(http.MethodPut, path, body, v)
}

func (c *Client) doJSON(method, path string, body any, v any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.checkAndDecode(resp, v)
}

// Delete sends a DELETE request to path.
func (c *Client) Delete(path string) error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return c.apiError(resp)
	}
	return nil
}

// PagedResponse is the Bitbucket paged list envelope.
type PagedResponse[T any] struct {
	Values        []T  `json:"values"`
	Size          int  `json:"size"`
	IsLastPage    bool `json:"isLastPage"`
	NextPageStart *int `json:"nextPageStart"`
	Start         int  `json:"start"`
}

func (c *Client) checkAndDecode(resp *http.Response, v any) error {
	if resp.StatusCode >= 400 {
		return c.apiError(resp)
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

type bbErrors struct {
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) apiError(resp *http.Response) error {
	var bbErr bbErrors
	_ = json.NewDecoder(resp.Body).Decode(&bbErr)
	msg := ""
	if len(bbErr.Errors) > 0 {
		msg = bbErr.Errors[0].Message
	}
	url := ""
	if resp.Request != nil {
		url = resp.Request.URL.String()
	}
	return &HTTPError{StatusCode: resp.StatusCode, Message: msg, RequestURL: url}
}
