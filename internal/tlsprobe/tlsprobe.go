// Package tlsprobe runs a handshake-only TLS dial against a host so
// callers (currently `bitbottle auth login`) can decide whether the
// host's certificate chains to a CA the OS already trusts. When the
// handshake fails specifically because of an unknown authority, the
// probe re-dials with InsecureSkipVerify and captures the leaf
// certificate — that cert is what the caller renders to the user so
// they can decide whether to trust it (SSH known-hosts UX).
//
// Network-class errors (DNS failure, connection refused, timeout) are
// returned as errors, not Results — the caller is then free to let
// the normal request path surface them with the usual error catalogue.
package tlsprobe

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// Options carry the knobs that tests need to override. Production
// callers pass the zero value to get sensible defaults.
type Options struct {
	// RootCAs overrides the system trust store. Tests use this to
	// distinguish "trusted CA" from "self-signed" without touching the
	// real OS keychain.
	RootCAs *x509.CertPool
	// Timeout caps the entire handshake. Zero → 10s.
	Timeout time.Duration
}

// Result describes the outcome of a successful or self-signed probe.
// A nil Result means the probe could not even reach the host (the
// returned error carries the network-class reason).
type Result struct {
	// TrustedByOS is true when the host's certificate chain verified
	// against the system trust store (or the override pool when
	// Options.RootCAs is non-nil). In that case LeafCert is nil — there
	// is nothing for the user to decide.
	TrustedByOS bool
	// LeafCert is populated only when TrustedByOS is false. It is the
	// leaf certificate returned by the server during the second,
	// InsecureSkipVerify dial — the cert the user would be trusting if
	// they confirmed.
	LeafCert *x509.Certificate
	// FingerprintSHA256 is the hex-encoded SHA-256 of LeafCert.Raw
	// (the DER-encoded certificate). Empty when LeafCert is nil. Use
	// this when printing to the user; it's the same value GitHub and
	// SSH known-hosts surface.
	FingerprintSHA256 string
}

// Probe performs a TLS handshake against host (which must already
// include a port, e.g. "git.example.com:443"). If host has no port
// the probe defaults to :443. The returned Result describes whether
// the OS trusted the chain. Network and context-cancellation errors
// short-circuit and are returned as errors.
func Probe(ctx context.Context, host string, opts Options) (*Result, error) {
	if host == "" {
		return nil, errors.New("tlsprobe: empty host")
	}
	if !strings.Contains(host, ":") {
		host += ":443"
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	// First dial: verification ON. The handshake either succeeds
	// (trusted) or fails. Distinguish "untrusted CA" from network/other
	// errors via errors.As against x509.UnknownAuthorityError and
	// x509.HostnameError. Anything else bubbles up.
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	d := &tls.Dialer{Config: &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    opts.RootCAs, // nil → system roots
		ServerName: hostNameOnly(host),
	}}
	conn, err := d.DialContext(dialCtx, "tcp", host)
	if err == nil {
		_ = conn.Close()
		return &Result{TrustedByOS: true}, nil
	}

	if !isUntrustedAuthorityError(err) {
		// Network failure, timeout, refused, DNS, etc — surface as-is.
		return nil, err
	}

	// Second dial: InsecureSkipVerify so the handshake completes and
	// the server hands us its certificate chain.
	dialCtx2, cancel2 := context.WithTimeout(ctx, timeout)
	defer cancel2()
	d2 := &tls.Dialer{Config: &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, //nolint:gosec // capturing the cert is the entire point
		ServerName:         hostNameOnly(host),
	}}
	conn2, err := d2.DialContext(dialCtx2, "tcp", host)
	if err != nil {
		return nil, fmt.Errorf("tlsprobe: capture-dial failed: %w", err)
	}
	defer func() { _ = conn2.Close() }()

	tlsConn, ok := conn2.(*tls.Conn)
	if !ok {
		return nil, fmt.Errorf("tlsprobe: unexpected conn type %T", conn2)
	}
	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, errors.New("tlsprobe: server returned no certificates")
	}
	leaf := certs[0]
	sum := sha256.Sum256(leaf.Raw)
	return &Result{
		TrustedByOS:       false,
		LeafCert:          leaf,
		FingerprintSHA256: hex.EncodeToString(sum[:]),
	}, nil
}

func isUntrustedAuthorityError(err error) bool {
	var uaErr x509.UnknownAuthorityError
	if errors.As(err, &uaErr) {
		return true
	}
	var hnErr x509.HostnameError
	if errors.As(err, &hnErr) {
		// Treat hostname mismatch the same way — the user is the only
		// one who can decide whether the cert presented is the one
		// they expect to see.
		return true
	}
	var certErr x509.CertificateInvalidError
	if errors.As(err, &certErr) {
		// Expired or otherwise-invalid corporate certs end up here
		// too. Self-signed certs without a chain frequently report
		// as Reason==NotAuthorizedToSign.
		return true
	}
	return false
}

func hostNameOnly(host string) string {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}
	return h
}
