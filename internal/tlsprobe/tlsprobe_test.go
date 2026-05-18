package tlsprobe_test

import (
	"context"
	"crypto/x509"
	"errors"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/tlsprobe"
)

func TestProbe_TrustedCA_ReturnsTrustedTrue(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(nil)
	t.Cleanup(srv.Close)

	// Trust httptest's self-signed cert via a custom rootCAs pool. This
	// is the "user already has the CA installed" path — Probe should
	// return TrustedByOS=true and no LeafCert.
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())

	host := mustHost(t, srv.URL)
	res, err := tlsprobe.Probe(context.Background(), host, tlsprobe.Options{
		RootCAs: pool,
		Timeout: 3 * time.Second,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.TrustedByOS, "expected trusted handshake against custom root pool")
	assert.Nil(t, res.LeafCert)
}

func TestProbe_SelfSigned_ReturnsLeafCert(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(nil)
	t.Cleanup(srv.Close)

	// Default system roots — httptest's CA is not installed, so the
	// handshake fails with x509.UnknownAuthorityError. Probe must
	// recognise that, dial a second time with InsecureSkipVerify, and
	// return the leaf cert so the caller can show it to the user.
	host := mustHost(t, srv.URL)
	res, err := tlsprobe.Probe(context.Background(), host, tlsprobe.Options{
		RootCAs: x509.NewCertPool(), // empty pool → guaranteed UnknownAuthority
		Timeout: 3 * time.Second,
	})
	require.NoError(t, err, "Probe should classify self-signed as a Result, not an error")
	require.NotNil(t, res)
	assert.False(t, res.TrustedByOS)
	require.NotNil(t, res.LeafCert, "expected leaf cert for self-signed handshake")
	assert.Equal(t, srv.Certificate().Subject.String(), res.LeafCert.Subject.String())
	assert.NotEmpty(t, res.FingerprintSHA256, "expected SHA-256 fingerprint")
	assert.Len(t, res.FingerprintSHA256, 64, "SHA-256 hex should be 64 chars")
}

func TestProbe_UnreachableHost_ReturnsError(t *testing.T) {
	t.Parallel()
	// Port 1 is reserved and effectively guaranteed to refuse on
	// loopback. Use a deliberately short timeout so the test stays
	// fast on platforms where the OS hangs the dial.
	res, err := tlsprobe.Probe(context.Background(), "127.0.0.1:1", tlsprobe.Options{
		Timeout: 500 * time.Millisecond,
	})
	require.Error(t, err, "network error must NOT be classified as a successful Result")
	assert.Nil(t, res)

	// Sanity: the error should be a network-class error, not a TLS one.
	var netErr net.Error
	if !errors.As(err, &netErr) && !strings.Contains(err.Error(), "refused") && !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected network-class error, got %T: %v", err, err)
	}
}

func TestProbe_ContextCancellation_ReturnsError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(nil)
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancelled

	res, err := tlsprobe.Probe(ctx, mustHost(t, srv.URL), tlsprobe.Options{
		Timeout: 3 * time.Second,
	})
	require.Error(t, err)
	assert.Nil(t, res)
}

func mustHost(t *testing.T, rawURL string) string {
	t.Helper()
	// httptest URLs look like https://127.0.0.1:54321 — strip the scheme.
	return strings.TrimPrefix(rawURL, "https://")
}
