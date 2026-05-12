// Package api re-exports the HTTP middleware stack from api/internal/httpx so
// callers outside the api/ subtree (notably pkg/cmd/factory) can wrap their
// http.Client transports without needing access to the internal package.
package api

import (
	"context"
	"net/http"

	"github.com/proggarapsody/bitbottle/api/internal/httpx"
)

// RetryPolicy is re-exported from api/internal/httpx so callers under pkg/cmd
// (which cannot import internal packages) can opt in to retry behaviour on a
// per-request basis via WithRetry.
type RetryPolicy = httpx.RetryPolicy

// DefaultRetryPolicy is 3 attempts, 200 ms → 2 s, with jitter.
var DefaultRetryPolicy = httpx.DefaultRetryPolicy

// WithRetry returns a new context carrying policy. Attach it to a request to
// enable retries through a transport produced by WrapTransport.
func WithRetry(ctx context.Context, policy RetryPolicy) context.Context {
	return httpx.WithRetry(ctx, policy)
}

// WrapTransport wraps inner with the httpx middleware stack (retry → ETag
// cache → rate limit). See api/internal/httpx for details.
func WrapTransport(inner http.RoundTripper) http.RoundTripper {
	return httpx.WrapTransport(inner)
}
