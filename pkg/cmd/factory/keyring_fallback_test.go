package factory

import (
	"testing"

	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestResolveToken_KeyringFallback guards Bug B from PRD #372: when
// hostCfg.OAuthToken is empty (the post-`auth migrate` shape, since
// migrate zeroes the field and config.MarshalYAML strips it on save),
// the factory must fall back to the keyring before constructing the
// backend client. Otherwise every API call after migration goes out
// with an empty token and 401s.
func TestResolveToken_KeyringFallback(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	if err := kr.Set("bitbottle", "alice", "kr-token"); err != nil {
		t.Fatalf("seed keyring: %v", err)
	}

	hc := config.HostConfig{User: "alice", OAuthToken: ""}
	got := resolveToken(hc, kr)
	if got != "kr-token" {
		t.Fatalf("resolveToken with empty OAuthToken: got %q, want %q", got, "kr-token")
	}
}

// TestResolveToken_PrefersConfig: if config has a token, use it (no
// keyring round-trip). This matches the pre-migrate behaviour and
// preserves the --with-token in-memory override path.
func TestResolveToken_PrefersConfig(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	_ = kr.Set("bitbottle", "alice", "kr-token")

	hc := config.HostConfig{User: "alice", OAuthToken: "cfg-token"}
	got := resolveToken(hc, kr)
	if got != "cfg-token" {
		t.Fatalf("resolveToken with non-empty OAuthToken: got %q, want %q", got, "cfg-token")
	}
}

// TestResolveToken_KeyringMiss_ReturnsEmpty: no token anywhere is not
// an error — the caller surfaces a clean "not authenticated" message.
func TestResolveToken_KeyringMiss_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	hc := config.HostConfig{User: "alice", OAuthToken: ""}
	got := resolveToken(hc, kr)
	if got != "" {
		t.Fatalf("resolveToken with empty config + empty keyring: got %q, want \"\"", got)
	}
}

// TestResolveToken_EmptyUser_NoLookup: with no User slug we cannot
// form the keyring key, so resolveToken must not attempt a lookup
// (returning empty cleanly rather than asking the OS keyring for "").
func TestResolveToken_EmptyUser_NoLookup(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	_ = kr.Set("bitbottle", "", "ghost-token")
	hc := config.HostConfig{User: "", OAuthToken: ""}
	got := resolveToken(hc, kr)
	if got != "" {
		t.Fatalf("resolveToken with empty User: got %q, want \"\"", got)
	}
}
