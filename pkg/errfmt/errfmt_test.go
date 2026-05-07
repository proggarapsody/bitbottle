package errfmt_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/errfmt"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// TestRender_Catalogue runs the central happy path for every code wired
// into the errfmt catalogue: each must produce a title line plus at
// least one hint line, with {{.Host}} substitution applied. New codes
// added to the catalogue should be added here too — the table is a
// living spec for what users see.
func TestRender_Catalogue(t *testing.T) {
	const host = "git.example.com"
	cases := []struct {
		name string
		code backend.ErrorCode
		// substrings that MUST appear on stderr in the order listed
		want []string
	}{
		{
			name: "auth.no_token",
			code: backend.CodeAuthNoToken,
			want: []string{
				"No credentials configured for git.example.com",
				"bitbottle auth login --hostname git.example.com",
			},
		},
		{
			name: "auth.invalid_token",
			code: backend.CodeAuthInvalidToken,
			want: []string{
				"Authentication failed on git.example.com",
				"bitbottle auth login --hostname git.example.com",
			},
		},
		{
			name: "perm.write_required",
			code: backend.CodePermWriteRequired,
			want: []string{
				"Permission denied on git.example.com",
				"lacks write scope",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ios := iostreams.Test()
			errfmt.Render(ios, &backend.DomainError{
				Code:    tc.code,
				Host:    host,
				Message: "underlying transport detail",
			})
			got := ios.ErrOut.(*bytes.Buffer).String()
			pos := 0
			for _, want := range tc.want {
				idx := strings.Index(got[pos:], want)
				if idx < 0 {
					t.Fatalf("missing %q in order\n--- got ---\n%s", want, got)
				}
				pos += idx + len(want)
			}
			if !strings.Contains(got, "Cause: underlying transport detail") {
				t.Errorf("expected Cause line carrying the underlying message\n--- got ---\n%s", got)
			}
		})
	}
}

// TestRender_NilError_NoOp verifies the renderer is safe to call from a
// generic error-handling shim that doesn't pre-check for nil.
func TestRender_NilError_NoOp(t *testing.T) {
	ios := iostreams.Test()
	errfmt.Render(ios, nil)
	if got := ios.ErrOut.(*bytes.Buffer).String(); got != "" {
		t.Errorf("expected no output for nil err, got %q", got)
	}
}

// TestRender_UnknownError_PassesThrough verifies non-DomainError values
// (and DomainErrors with no recognised code) fall back to the raw error
// string so the renderer is never worse than the legacy printer.
func TestRender_UnknownError_PassesThrough(t *testing.T) {
	ios := iostreams.Test()
	errfmt.Render(ios, errors.New("connection reset by peer"))
	got := ios.ErrOut.(*bytes.Buffer).String()
	if !strings.Contains(got, "connection reset by peer") {
		t.Errorf("expected raw error to pass through, got %q", got)
	}
}
