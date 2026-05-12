package httpx

import (
	"context"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// RateLimitState holds parsed X-RateLimit-* header values.
type RateLimitState struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// parseRateLimit reads X-RateLimit-{Limit,Remaining,Reset} from h.
// Returns zero RateLimitState and false if headers are absent.
func parseRateLimit(h http.Header) (RateLimitState, bool) {
	limit := h.Get("X-RateLimit-Limit")
	remaining := h.Get("X-RateLimit-Remaining")
	reset := h.Get("X-RateLimit-Reset")
	if limit == "" && remaining == "" && reset == "" {
		return RateLimitState{}, false
	}
	st := RateLimitState{}
	if n, err := strconv.Atoi(limit); err == nil {
		st.Limit = n
	}
	if n, err := strconv.Atoi(remaining); err == nil {
		st.Remaining = n
	}
	if n, err := strconv.ParseInt(reset, 10, 64); err == nil {
		st.Reset = time.Unix(n, 0)
	}
	return st, true
}

// rateLimitRoundTripper wraps inner and reads rate-limit headers from each
// response. When Remaining == 0, it blocks until Reset + per-call jitter
// (0–500 ms) before returning the response to the caller. If headers are
// absent, it is a no-op.
type rateLimitRoundTripper struct {
	inner   http.RoundTripper
	nowFn   func() time.Time
	sleepFn func(time.Duration)
}

func (r *rateLimitRoundTripper) now() time.Time {
	if r.nowFn != nil {
		return r.nowFn()
	}
	return time.Now()
}

func (r *rateLimitRoundTripper) sleep(ctx context.Context, d time.Duration) error {
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

func (r *rateLimitRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := r.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	st, ok := parseRateLimit(resp.Header)
	if !ok {
		return resp, nil
	}
	if st.Remaining > 0 {
		return resp, nil
	}
	if st.Reset.IsZero() {
		return resp, nil
	}
	wait := st.Reset.Sub(r.now())
	if wait <= 0 {
		return resp, nil
	}
	jitter := time.Duration(rand.Int63n(int64(500 * time.Millisecond)))
	if err := r.sleep(req.Context(), wait+jitter); err != nil {
		return nil, err
	}
	return resp, nil
}
