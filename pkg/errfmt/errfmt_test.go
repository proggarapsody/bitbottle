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
		// optional context fields stamped onto the DomainError so codes
		// whose templates reference {{.ID}} / {{.Resource}} / {{.Feature}}
		// can be exercised end-to-end.
		id       string
		resource string
		feature  string
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
		{
			name:     "repo.not_found",
			code:     backend.CodeRepoNotFound,
			id:       "ws/repo",
			resource: "repository",
			want: []string{
				"Repository ws/repo not found on git.example.com",
				"Check the slug casing",
			},
		},
		{
			name:     "pr.not_found",
			code:     backend.CodePRNotFound,
			id:       "42",
			resource: "pull-request",
			want: []string{
				"Pull request #42 not found on git.example.com",
				"`bitbottle pr list`",
			},
		},
		{
			name:     "pr.merge.conflict",
			code:     backend.CodePRMergeConflict,
			id:       "42",
			resource: "pull-request",
			want: []string{
				"Pull request #42 cannot be merged",
				"Resolve conflicts locally",
			},
		},
		{
			name:     "pr.merge.behind",
			code:     backend.CodePRMergeBehind,
			id:       "42",
			resource: "pull-request",
			want: []string{
				"behind its target branch on git.example.com",
				"Update the source branch",
			},
		},
		{
			name: "pr.create.duplicate_branch",
			code: backend.CodePRCreateDuplicateBranch,
			want: []string{
				"A pull request for these branches already exists",
				"close it before creating",
			},
		},
		{
			name: "pr.reviewer.unknown",
			code: backend.CodePRReviewerUnknown,
			want: []string{
				"One or more reviewers are not members of git.example.com",
				"Check the slug spelling",
			},
		},
		{
			name: "branch.protected",
			code: backend.CodeBranchProtected,
			want: []string{
				"Branch is protected on git.example.com",
				"`bitbottle branch protect list`",
			},
		},
		{
			name:    "host.unsupported",
			code:    backend.CodeHostUnsupported,
			feature: "fork",
			want: []string{
				"`fork` is not available on git.example.com",
				"different Bitbucket flavour",
			},
		},
		{
			name: "network.tls_unknown_authority",
			code: backend.CodeNetworkTLSUnknownAuthority,
			want: []string{
				"TLS verification failed for git.example.com",
				"`-k`",
				"`skip_tls_verify: true`",
			},
		},
		{
			name: "transport.timeout",
			code: backend.CodeTransportTimeout,
			want: []string{
				"Request to git.example.com timed out",
				"Retry in a moment",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ios := iostreams.Test()
			errfmt.Render(ios, &backend.DomainError{
				Code:     tc.code,
				Host:     host,
				ID:       tc.id,
				Resource: tc.resource,
				Feature:  tc.feature,
				Message:  "underlying transport detail",
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

// TestCatalogue_CoversAllPublishedCodes is the drift gate: every code
// listed in backend.AllCodes must produce non-empty stderr output via
// the catalogue path. Adding a new ErrorCode without a matching errfmt
// entry will fall through to renderByKind → raw err.Error(); this test
// fails loudly when that happens.
//
// Workflow when adding a code:
//  1. Append it to backend.AllCodes in api/backend/errors.go.
//  2. Add a catalogue entry in pkg/errfmt/errfmt.go.
//  3. This test passes.
func TestCatalogue_CoversAllPublishedCodes(t *testing.T) {
	if len(backend.AllCodes) == 0 {
		t.Fatal("backend.AllCodes is empty — catalogue gate cannot run")
	}
	for _, code := range backend.AllCodes {
		t.Run(string(code), func(t *testing.T) {
			ios := iostreams.Test()
			errfmt.Render(ios, &backend.DomainError{
				Code: code,
				Host: "h.example",
			})
			got := ios.ErrOut.(*bytes.Buffer).String()
			if got == "" {
				t.Errorf("code %q produced no output — missing errfmt catalogue entry", code)
			}
			// Every existing template includes {{.Host}}; if a future code
			// legitimately omits it, relax this check.
			if !strings.Contains(got, "h.example") {
				t.Errorf("code %q did not appear to use the catalogue path: %q", code, got)
			}
		})
	}
}

// TestRender_KindFallback_BareForms pins the strings produced when a
// renderByKind arm runs with no contextual fields populated. These are
// the safety-net lines users see when an adapter classified the error
// but didn't add Resource/ID/Feature/Host yet — they must remain stable
// so future refactors don't silently change the user-visible text.
func TestRender_KindFallback_BareForms(t *testing.T) {
	cases := []struct {
		name string
		err  *backend.DomainError
		want string
	}{
		{
			name: "permission no host",
			err:  &backend.DomainError{Kind: backend.ErrPermission},
			want: "Permission denied. Check your token scopes or re-run `bitbottle auth login`.",
		},
		{
			name: "not-found no resource or id",
			err:  &backend.DomainError{Kind: backend.ErrNotFound},
			want: "Resource not found.",
		},
		{
			name: "unsupported no feature",
			err:  &backend.DomainError{Kind: backend.ErrUnsupportedOnHost},
			want: "this feature is not available.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ios := iostreams.Test()
			errfmt.Render(ios, tc.err)
			got := ios.ErrOut.(*bytes.Buffer).String()
			if !strings.Contains(got, tc.want) {
				t.Errorf("expected %q in output, got %q", tc.want, got)
			}
		})
	}
}

// TestRender_CodePrecedence pins the resolution order between Code and
// Kind: when both are set, the catalogue (Code) wins; when Code is set
// but unknown, the renderer falls back to the Kind path so an in-flight
// rename or typo doesn't blank out user feedback.
func TestRender_CodePrecedence(t *testing.T) {
	t.Run("known code beats kind", func(t *testing.T) {
		ios := iostreams.Test()
		errfmt.Render(ios, &backend.DomainError{
			Code: backend.CodeAuthInvalidToken,
			Kind: backend.ErrConflict,
			Host: "h.example",
		})
		got := ios.ErrOut.(*bytes.Buffer).String()
		if !strings.Contains(got, "Authentication failed") {
			t.Errorf("catalogue path should win, got %q", got)
		}
		if strings.Contains(got, "already exists") {
			t.Errorf("Kind-based conflict line should not appear when Code is recognised, got %q", got)
		}
	})
	t.Run("unknown code falls through to kind", func(t *testing.T) {
		ios := iostreams.Test()
		errfmt.Render(ios, &backend.DomainError{
			Code: backend.ErrorCode("made.up.code"),
			Kind: backend.ErrConflict,
		})
		got := ios.ErrOut.(*bytes.Buffer).String()
		if !strings.Contains(got, "already exists") {
			t.Errorf("unknown code should fall through to Kind path, got %q", got)
		}
	})
}
