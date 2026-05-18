package auth

import "testing"

// TestNormalizeHostname pins behaviour of normalizeHostname: strips
// http://, https:// prefix and a trailing slash so downstream URL
// builders that do fmt.Sprintf("https://%s/...", host) cannot produce
// "https://https://HOST/..." (Bug A in PRD #372).
func TestNormalizeHostname(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"bare host", "git.moscow.alfaintra.net", "git.moscow.alfaintra.net"},
		{"https prefix", "https://git.moscow.alfaintra.net", "git.moscow.alfaintra.net"},
		{"http prefix", "http://example.com", "example.com"},
		{"https with trailing slash", "https://bitbucket.org/", "bitbucket.org"},
		{"http with trailing slash", "http://example.com/", "example.com"},
		{"bare with trailing slash", "example.com/", "example.com"},
		{"mixed-case https", "HTTPS://Example.COM", "Example.COM"},
		{"mixed-case http", "HtTp://Example.com/", "Example.com"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeHostname(tc.in)
			if got != tc.want {
				t.Fatalf("normalizeHostname(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
