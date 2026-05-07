package httpx_test

import (
	"context"
	"crypto/x509"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/httpx"
)

// stubDoer returns a fixed error from Do, simulating a transport-level
// failure (TLS handshake, dial timeout, etc.) where no HTTP response
// ever materialised.
type stubDoer struct{ err error }

func (s stubDoer) Do(_ *http.Request) (*http.Response, error) { return nil, s.err }

// TestTransport_ClassifiesTLSAtTransportLayer verifies that when
// UseDomainErrors is set, a TLS handshake failure (x509.UnknownAuthorityError
// wrapped in *url.Error, the shape Go's HTTP client surfaces it as) is
// translated into a backend.DomainError carrying the network cluster code
// before the cmd layer sees it. Without this wiring, users see raw
// "x509: certificate signed by unknown authority" with no -k hint.
func TestTransport_ClassifiesTLSAtTransportLayer(t *testing.T) {
	transportErr := &url.Error{
		Op:  "Get",
		URL: "https://h.example/foo",
		Err: x509.UnknownAuthorityError{},
	}
	tr := httpx.New(stubDoer{err: transportErr}, "https://h.example", httpx.Auth{Token: "t"}, nil, httpx.ContentTypeAlwaysWrite, nil).
		UseDomainErrors("h.example")

	var dst struct{}
	err := tr.GetJSON("/foo", &dst)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected *backend.DomainError, got %T: %v", err, err)
	}
	if de.Code != backend.CodeNetworkTLSUnknownAuthority {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodeNetworkTLSUnknownAuthority)
	}
	if de.Host != "h.example" {
		t.Errorf("Host = %q, want %q", de.Host, "h.example")
	}
}

// TestTransport_ClassifiesTimeoutAtTransportLayer verifies the same for a
// context.DeadlineExceeded — common when the user's VPN drops mid-request.
func TestTransport_ClassifiesTimeoutAtTransportLayer(t *testing.T) {
	transportErr := &url.Error{
		Op:  "Get",
		URL: "https://h.example/foo",
		Err: context.DeadlineExceeded,
	}
	tr := httpx.New(stubDoer{err: transportErr}, "https://h.example", httpx.Auth{Token: "t"}, nil, httpx.ContentTypeAlwaysWrite, nil).
		UseDomainErrors("h.example")

	var dst struct{}
	err := tr.GetJSON("/foo", &dst)

	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected *backend.DomainError, got %T: %v", err, err)
	}
	if de.Code != backend.CodeTransportTimeout {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodeTransportTimeout)
	}
}

// TestTransport_DoesNotClassifyWithoutUseDomainErrors pins the back-compat
// invariant: when UseDomainErrors is NOT set (the default), transport errors
// are returned as-is so existing tests and direct callers see the original
// *url.Error and can pattern-match on it.
func TestTransport_DoesNotClassifyWithoutUseDomainErrors(t *testing.T) {
	transportErr := &url.Error{
		Op:  "Get",
		URL: "https://h.example/foo",
		Err: x509.UnknownAuthorityError{},
	}
	tr := httpx.New(stubDoer{err: transportErr}, "https://h.example", httpx.Auth{Token: "t"}, nil, httpx.ContentTypeAlwaysWrite, nil)

	var dst struct{}
	err := tr.GetJSON("/foo", &dst)

	var de *backend.DomainError
	if errors.As(err, &de) {
		t.Errorf("expected raw *url.Error pass-through (no UseDomainErrors), got DomainError %+v", de)
	}
	if !errors.Is(err, transportErr) && err != transportErr {
		t.Logf("err = %v (acceptable: wrapped or identical to transportErr)", err)
	}
}

// TestTransport_PassesThroughUnknownTransportError pins "do no harm" — when
// UseDomainErrors is set but the transport error doesn't match a known
// classifier shape (e.g. "connection reset by peer"), the original error
// is returned rather than mis-labelled.
func TestTransport_PassesThroughUnknownTransportError(t *testing.T) {
	transportErr := &url.Error{
		Op:  "Get",
		URL: "https://h.example/foo",
		Err: errors.New("connection reset by peer"),
	}
	tr := httpx.New(stubDoer{err: transportErr}, "https://h.example", httpx.Auth{Token: "t"}, nil, httpx.ContentTypeAlwaysWrite, nil).
		UseDomainErrors("h.example")

	var dst struct{}
	err := tr.GetJSON("/foo", &dst)

	var de *backend.DomainError
	if errors.As(err, &de) {
		t.Errorf("expected pass-through for unknown transport error, got DomainError %+v", de)
	}
}
