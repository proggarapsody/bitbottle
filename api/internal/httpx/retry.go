package httpx

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// RetryPolicy controls retry behaviour for a single request.
type RetryPolicy struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Jitter         bool
}

// DefaultRetryPolicy is 3 attempts, 200 ms → 2 s, with jitter.
var DefaultRetryPolicy = RetryPolicy{
	MaxAttempts:    3,
	InitialBackoff: 200 * time.Millisecond,
	MaxBackoff:     2 * time.Second,
	Jitter:         true,
}

type retryContextKey struct{}

// WithRetry returns a new context carrying policy. Pass it to any transport
// method (GetJSON, PostJSON, etc.) to enable retries for that call.
func WithRetry(ctx context.Context, policy RetryPolicy) context.Context {
	return context.WithValue(ctx, retryContextKey{}, policy)
}

// retryRoundTripper wraps inner and retries on 5xx / 429.
// Opt-in: only retries when the request context contains a RetryPolicy.
// On 429 with Retry-After: use that delay instead of backoff.
// 4xx (except 429) are never retried.
// req.GetBody must be non-nil for any request with a body; panics in dev if not.
type retryRoundTripper struct {
	inner   http.RoundTripper
	sleepFn func(time.Duration)
}

func (r *retryRoundTripper) sleep(ctx context.Context, d time.Duration) error {
	if r.sleepFn != nil {
		r.sleepFn(d)
		return nil
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	policy, ok := req.Context().Value(retryContextKey{}).(RetryPolicy)
	if !ok {
		return r.inner.RoundTrip(req)
	}
	if req.Body != nil && req.GetBody == nil {
		panic("httpx: retry policy set on request with body but GetBody is nil")
	}
	var lastResp *http.Response
	var lastErr error
	maxAttempts := policy.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 && req.Body != nil && req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				return nil, err
			}
			req.Body = body
		}
		resp, err := r.inner.RoundTrip(req)
		lastResp = resp
		lastErr = err
		if err != nil {
			if attempt+1 >= maxAttempts {
				return nil, err
			}
			if sleepErr := r.sleep(req.Context(), computeBackoff(policy, attempt)); sleepErr != nil {
				return nil, sleepErr
			}
			continue
		}
		if !shouldRetryStatus(resp.StatusCode) {
			return resp, nil
		}
		if attempt+1 >= maxAttempts {
			return resp, nil
		}
		_ = resp.Body.Close()
		delay := computeBackoff(policy, attempt)
		if resp.StatusCode == http.StatusTooManyRequests {
			if ra := parseRetryAfter(resp.Header.Get("Retry-After")); ra > 0 {
				delay = ra
			}
		}
		if sleepErr := r.sleep(req.Context(), delay); sleepErr != nil {
			return nil, sleepErr
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return lastResp, nil
}

func shouldRetryStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

func computeBackoff(p RetryPolicy, attempt int) time.Duration {
	if p.InitialBackoff <= 0 {
		return 0
	}
	backoff := p.InitialBackoff
	for i := 0; i < attempt; i++ {
		backoff *= 2
		if p.MaxBackoff > 0 && backoff > p.MaxBackoff {
			backoff = p.MaxBackoff
			break
		}
	}
	if p.MaxBackoff > 0 && backoff > p.MaxBackoff {
		backoff = p.MaxBackoff
	}
	if p.Jitter && backoff > 0 {
		backoff = time.Duration(rand.Int63n(int64(backoff)))
	}
	return backoff
}

func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		return time.Until(t)
	}
	return 0
}
