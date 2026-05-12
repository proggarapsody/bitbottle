package httpx

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRateLimit_Present(t *testing.T) {
	reset := time.Now().Add(30 * time.Second).Unix()
	h := http.Header{}
	h.Set("X-RateLimit-Limit", "100")
	h.Set("X-RateLimit-Remaining", "42")
	h.Set("X-RateLimit-Reset", asString(reset))
	st, ok := parseRateLimit(h)
	require.True(t, ok)
	assert.Equal(t, 100, st.Limit)
	assert.Equal(t, 42, st.Remaining)
	assert.Equal(t, reset, st.Reset.Unix())
}

func TestParseRateLimit_Absent(t *testing.T) {
	_, ok := parseRateLimit(http.Header{})
	assert.False(t, ok)
}

func TestRateLimitRoundTripper_NoHeaders_NoSleep(t *testing.T) {
	inner := &fakeRoundTripper{responses: []*http.Response{makeResp(200, nil)}} //nolint:bodyclose
	called := false
	rt := &rateLimitRoundTripper{
		inner:   inner,
		sleepFn: func(time.Duration) { called = true },
		nowFn:   time.Now,
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/x", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.Equal(t, 200, resp.StatusCode)
	assert.False(t, called)
}

func TestRateLimitRoundTripper_RemainingGTZero_NoSleep(t *testing.T) {
	inner := &fakeRoundTripper{responses: []*http.Response{makeResp(200, map[string]string{ //nolint:bodyclose
		"X-RateLimit-Limit":     "100",
		"X-RateLimit-Remaining": "5",
		"X-RateLimit-Reset":     asString(time.Now().Add(10 * time.Second).Unix()),
	})}}
	called := false
	rt := &rateLimitRoundTripper{
		inner:   inner,
		sleepFn: func(time.Duration) { called = true },
		nowFn:   time.Now,
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/x", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	assert.False(t, called)
}

func TestRateLimitRoundTripper_Exhausted_SleepsUntilReset(t *testing.T) {
	// truncate now to whole seconds so Reset header round-trip is lossless
	now := time.Unix(time.Now().Unix(), 0)
	reset := now.Add(1 * time.Second)
	inner := &fakeRoundTripper{responses: []*http.Response{{
		StatusCode: 200,
		Header: http.Header{
			"X-Ratelimit-Limit":     []string{"100"},
			"X-Ratelimit-Remaining": []string{"0"},
			"X-Ratelimit-Reset":     []string{asString(reset.Unix())},
		},
		Body: io.NopCloser(strings.NewReader("")),
	}}}
	var slept []time.Duration
	rt := &rateLimitRoundTripper{
		inner:   inner,
		sleepFn: func(d time.Duration) { slept = append(slept, d) },
		nowFn:   func() time.Time { return now },
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/x", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Len(t, slept, 1)
	assert.GreaterOrEqual(t, slept[0], 1*time.Second)
	assert.LessOrEqual(t, slept[0], 1*time.Second+600*time.Millisecond)
}

func asString(n int64) string {
	return strings.TrimSpace(intToStr(n))
}

func intToStr(n int64) string {
	// avoid strconv import-bloat in tests
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
