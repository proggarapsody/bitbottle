package backend

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
)

// ClassifyTransportError translates a transport-level error (one that
// failed before an HTTP response existed — TLS handshake failure, DNS
// resolution timeout, dial timeout, read timeout, etc.) into a
// DomainError carrying a network-cluster ErrorCode. host is attached so
// the renderer can name which Bitbucket instance produced the failure.
//
// Returns nil if err is nil, or if err is not a recognised transport
// failure shape — callers should treat the original error as
// untranslated and either return it as-is or fall through to a generic
// renderer. The intent is "do no harm": never mis-label an error.
//
// Recognised shapes:
//   - x509.UnknownAuthorityError (anywhere in the unwrap chain) →
//     CodeNetworkTLSUnknownAuthority. Triggered by self-signed CAs;
//     the catalogue hints at -k / skip_tls_verify.
//   - context.DeadlineExceeded (anywhere in the unwrap chain) →
//     CodeTransportTimeout. Triggered by per-request deadlines.
//   - any net.Error whose Timeout() returns true →
//     CodeTransportTimeout. Covers dial timeouts and read timeouts
//     surfaced as *net.OpError.
//
// Kind is set to ErrTransport in every classified case so callers can
// branch on errors.Is(err, backend.ErrTransport).
func ClassifyTransportError(host string, err error) *DomainError {
	if err == nil {
		return nil
	}
	de := &DomainError{
		Host:    host,
		Cause:   err,
		Message: err.Error(),
		Kind:    ErrTransport,
	}
	var unkAuth x509.UnknownAuthorityError
	if errors.As(err, &unkAuth) {
		de.Code = CodeNetworkTLSUnknownAuthority
		return de
	}
	if errors.Is(err, context.DeadlineExceeded) {
		de.Code = CodeTransportTimeout
		return de
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		de.Code = CodeTransportTimeout
		return de
	}
	return nil
}
