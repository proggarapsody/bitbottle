package acceptancecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const cloudAPI = "https://api.bitbucket.org/2.0"

type apiClient struct {
	base, token string
	http        *http.Client
}

func newAPI(base, token string) (*apiClient, error) {
	if base == "" {
		base = cloudAPI
	}
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return nil, errors.New("invalid API base URL")
	}
	return &apiClient{base: strings.TrimRight(base, "/"), token: token, http: &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("API redirect refused") }}}, nil
}

func (c *apiClient) request(ctx context.Context, method, path string, body io.Reader, contentType string, out any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cloud API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if out != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(out); err != nil {
			return resp.StatusCode, errors.New("cloud API returned invalid JSON")
		}
	} else {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("cloud API returned HTTP %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func (c *apiClient) json(ctx context.Context, method, path string, input, output any) (int, error) {
	var body io.Reader
	if input != nil {
		b, err := json.Marshal(input)
		if err != nil {
			return 0, err
		}
		body = bytes.NewReader(b)
	}
	return c.request(ctx, method, path, body, "application/json", output)
}

func (c *apiClient) upload(ctx context.Context, path, branch, message, filename, content string) error {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("branch", branch)
	_ = w.WriteField("message", message)
	f, err := w.CreateFormFile(filename, filename)
	if err != nil {
		return err
	}
	_, _ = io.WriteString(f, content)
	if err = w.Close(); err != nil {
		return err
	}
	_, err = c.request(ctx, http.MethodPost, path, &body, w.FormDataContentType(), nil)
	return err
}

func esc(s string) string { return url.PathEscape(s) }
