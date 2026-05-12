package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRoundTripper returns the queued responses in order. After the queue is
// exhausted, it errors. If errs[i] is non-nil, the response is not consulted.
type fakeRoundTripper struct {
	responses []*http.Response
	errs      []error
	calls     int32
	lastReq   *http.Request
}

func (f *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	idx := atomic.AddInt32(&f.calls, 1) - 1
	f.lastReq = req
	if int(idx) >= len(f.responses) && int(idx) >= len(f.errs) {
		return nil, fmt.Errorf("no more queued responses (call %d)", idx)
	}
	if int(idx) < len(f.errs) && f.errs[int(idx)] != nil {
		return nil, f.errs[int(idx)]
	}
	if int(idx) < len(f.responses) {
		return f.responses[int(idx)], nil
	}
	return nil, fmt.Errorf("no response queued")
}

func makeResp(status int, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

func TestRetryRoundTripper_NoPolicy_NoRetry(t *testing.T) {
	inner := &fakeRoundTripper{
		responses: []*http.Response{makeResp(500, nil)}, //nolint:bodyclose
	}
	rt := &retryRoundTripper{inner: inner, sleepFn: func(time.Duration) {}}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/x", nil)
	require.NoError(t, err)
	resp, reqErr := rt.RoundTrip(req)
	require.NoError(t, reqErr)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, int32(1), inner.calls)
}

func TestRetryRoundTripper_5xx_Retries(t *testing.T) {
	inner := &fakeRoundTripper{
		responses: []*http.Response{
			makeResp(500, nil), //nolint:bodyclose // closed by retryRoundTripper before next attempt
			makeResp(500, nil), //nolint:bodyclose // closed by retryRoundTripper before next attempt
			makeResp(200, nil), //nolint:bodyclose // closed by caller after RoundTrip returns
		},
	}
	rt := &retryRoundTripper{inner: inner, sleepFn: func(time.Duration) {}}
	ctx := WithRetry(context.Background(), RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/x", nil)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, int32(3), inner.calls)
}

func TestRetryRoundTripper_429_RespectsRetryAfter(t *testing.T) {
	inner := &fakeRoundTripper{
		responses: []*http.Response{
			makeResp(429, map[string]string{"Retry-After": "2"}), //nolint:bodyclose
			makeResp(200, nil), //nolint:bodyclose
		},
	}
	var slept []time.Duration
	rt := &retryRoundTripper{
		inner:   inner,
		sleepFn: func(d time.Duration) { slept = append(slept, d) },
	}
	ctx := WithRetry(context.Background(), RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/x", nil)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, 200, resp.StatusCode)
	require.Len(t, slept, 1)
	assert.GreaterOrEqual(t, slept[0], 2*time.Second)
}

func TestRetryRoundTripper_4xx_NoRetry(t *testing.T) {
	inner := &fakeRoundTripper{
		responses: []*http.Response{makeResp(400, nil)}, //nolint:bodyclose
	}
	rt := &retryRoundTripper{inner: inner, sleepFn: func(time.Duration) {}}
	ctx := WithRetry(context.Background(), RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/x", nil)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, int32(1), inner.calls)
}

func TestRetryRoundTripper_NetworkError_Retries(t *testing.T) {
	netErr := errors.New("connection refused")
	inner := &fakeRoundTripper{
		errs: []error{netErr, netErr, netErr},
	}
	rt := &retryRoundTripper{inner: inner, sleepFn: func(time.Duration) {}}
	ctx := WithRetry(context.Background(), RetryPolicy{
		MaxAttempts:    3,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/x", nil)
	_, err := rt.RoundTrip(req) //nolint:bodyclose // error path returns nil response
	require.Error(t, err)
	assert.Equal(t, int32(3), inner.calls)
}

func TestWithRetry_ContextPropagation(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: 5, InitialBackoff: time.Second}
	ctx := WithRetry(context.Background(), policy)
	got, ok := ctx.Value(retryContextKey{}).(RetryPolicy)
	require.True(t, ok)
	assert.Equal(t, 5, got.MaxAttempts)
	assert.Equal(t, time.Second, got.InitialBackoff)
}
