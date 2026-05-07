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

// TestRender_StripsControlBytes is the regression guard for the terminal
// escape injection class (CWE-150 / CWE-117). A hostile or compromised
// Bitbucket Server can return error bodies containing CSI sequences,
// OSC 8 hyperlinks, OSC 52 clipboard writes, BEL bytes, or bracketed-paste
// escapes. Every server- or config-derived field that flows into stderr
// (Message, Host, Resource, ID, Feature) must be stripped of C0/C1
// control bytes and lone ESC before printing — newline and tab survive
// because they're load-bearing in legitimate output.
func TestRender_StripsControlBytes(t *testing.T) {
	cases := []struct {
		name string
		err  *backend.DomainError
	}{
		{
			name: "catalogue path: Message + Host carry escapes",
			err: &backend.DomainError{
				Code:    backend.CodeAuthInvalidToken,
				Host:    "h.example\x1b[31m",
				Message: "denied\x1b[2J\x07\x1b]8;;evil://x\x07click\x1b]8;;\x07",
			},
		},
		{
			name: "renderByKind path: Permission with escape in Host",
			err: &backend.DomainError{
				Kind: backend.ErrPermission,
				Host: "h.example\x1b[31mFAKE\x1b[0m",
			},
		},
		{
			name: "renderByKind path: NotFound with escape in Resource/ID",
			err: &backend.DomainError{
				Kind:     backend.ErrNotFound,
				Resource: "pull-request\x1b[2J",
				ID:       "42\x07",
			},
		},
		{
			name: "renderByKind path: UnsupportedOnHost with escape in Feature",
			err: &backend.DomainError{
				Kind:    backend.ErrUnsupportedOnHost,
				Feature: "fork\x1b]52;c;ZWlk\x07",
				Host:    "h.example",
			},
		},
		{
			name: "raw passthrough: plain error with escape in message",
			err: &backend.DomainError{
				Message: "bare\x1b[31mFAKE\x1b[0m\x07",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ios := iostreams.Test()
			errfmt.Render(ios, tc.err)
			got := ios.ErrOut.(*bytes.Buffer).String()
			for _, bad := range []string{"\x1b", "\x07", "\x00", "\x9b"} {
				if strings.Contains(got, bad) {
					t.Errorf("control byte %q reached stderr: %q", bad, got)
				}
			}
		})
	}
}

// TestRender_PreservesNewlineAndTab pins the carve-out: \n and \t are
// load-bearing (newlines split title/cause/hint lines, tabs sometimes
// appear in legitimate upstream messages) and must survive the sanitiser.
func TestRender_PreservesNewlineAndTab(t *testing.T) {
	ios := iostreams.Test()
	errfmt.Render(ios, &backend.DomainError{
		Code:    backend.CodeAuthInvalidToken,
		Host:    "h.example",
		Message: "line one\n\tindented detail",
	})
	got := ios.ErrOut.(*bytes.Buffer).String()
	if !strings.Contains(got, "line one\n\tindented detail") {
		t.Errorf("expected newline + tab preserved in Cause, got %q", got)
	}
}
