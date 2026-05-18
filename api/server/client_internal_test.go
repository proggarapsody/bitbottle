package server

import "testing"

// TestHostFromURLBare pins that hostFromURL returns the bare hostname
// (no scheme), so that DomainError.Host renders correctly in errfmt
// hints — e.g. `--hostname git.example.com`, not
// `--hostname https://git.example.com` which trips the scheme-stripping
// path in `auth login` (Bug D in PRD #372).
func TestHostFromURLBare(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"https://git.example.com/rest/api/1.0", "git.example.com"},
		{"http://internal.example.com/rest/api/1.0", "internal.example.com"},
		{"https://git.moscow.alfaintra.net/rest/api/1.0", "git.moscow.alfaintra.net"},
	}
	for _, tc := range cases {
		got := hostFromURL(tc.in)
		if got != tc.want {
			t.Errorf("hostFromURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestHostFromURL_Unparseable returns input untouched.
func TestHostFromURL_Unparseable(t *testing.T) {
	t.Parallel()
	got := hostFromURL("::::not-a-url::::")
	if got != "::::not-a-url::::" {
		t.Errorf("hostFromURL on unparseable input returned %q", got)
	}
}
