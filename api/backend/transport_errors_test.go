package backend_test

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"net/url"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// timeoutErr is a synthetic net.Error that reports Timeout()=true.
// We can't easily construct a real *net.OpError without a live socket,
// so we exercise the classifier's net.Error/Timeout path with a stub.
type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

func TestClassifyTransportError_RecognisesTLSUnknownAuthority(t *testing.T) {
	// Bitbucket Server with self-signed CA: Go's TLS handshake surfaces
	// x509.UnknownAuthorityError wrapped in *url.Error. The classifier
	// must turn it into network.tls_unknown_authority so errfmt can hint
	// the user toward -k / skip_tls_verify.
	wrapped := &url.Error{
		Op:  "Get",
		URL: "https://h.example/foo",
		Err: x509.UnknownAuthorityError{},
	}
	de := backend.ClassifyTransportError("h.example", wrapped)
	if de == nil {
		t.Fatal("expected DomainError, got nil")
	}
	if de.Code != backend.CodeNetworkTLSUnknownAuthority {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodeNetworkTLSUnknownAuthority)
	}
	if de.Host != "h.example" {
		t.Errorf("Host = %q, want %q", de.Host, "h.example")
	}
	if !errors.Is(de, backend.ErrTransport) {
		t.Errorf("expected Kind to match ErrTransport")
	}
}

func TestClassifyTransportError_RecognisesContextDeadline(t *testing.T) {
	// HTTP client with a per-request deadline that fires returns
	// *url.Error wrapping context.DeadlineExceeded.
	wrapped := &url.Error{
		Op:  "Get",
		URL: "https://h.example/foo",
		Err: context.DeadlineExceeded,
	}
	de := backend.ClassifyTransportError("h.example", wrapped)
	if de == nil {
		t.Fatal("expected DomainError, got nil")
	}
	if de.Code != backend.CodeTransportTimeout {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodeTransportTimeout)
	}
	if !errors.Is(de, backend.ErrTransport) {
		t.Errorf("expected Kind to match ErrTransport")
	}
}

func TestClassifyTransportError_RecognisesNetTimeout(t *testing.T) {
	// Lower-level net.OpError shapes (dial timeout, read timeout) implement
	// net.Error with Timeout()=true. The classifier must treat them the
	// same as context.DeadlineExceeded so users see a consistent hint.
	var netErr net.Error = timeoutErr{}
	wrapped := &url.Error{
		Op:  "Post",
		URL: "https://h.example/foo",
		Err: netErr,
	}
	de := backend.ClassifyTransportError("h.example", wrapped)
	if de == nil {
		t.Fatal("expected DomainError, got nil")
	}
	if de.Code != backend.CodeTransportTimeout {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodeTransportTimeout)
	}
}

func TestClassifyTransportError_PassesThroughUnknown(t *testing.T) {
	// Errors we can't classify (DNS resolution failure with a non-timeout
	// shape, miscellaneous dial errors) should return nil so the caller
	// keeps the original error rather than mis-labelling it.
	wrapped := &url.Error{
		Op:  "Get",
		URL: "https://h.example/foo",
		Err: errors.New("some unrecognised transport error"),
	}
	if de := backend.ClassifyTransportError("h.example", wrapped); de != nil {
		t.Errorf("expected nil for unknown error, got %+v", de)
	}
}

func TestClassifyTransportError_NilErrorReturnsNil(t *testing.T) {
	if de := backend.ClassifyTransportError("h.example", nil); de != nil {
		t.Errorf("expected nil, got %+v", de)
	}
}
