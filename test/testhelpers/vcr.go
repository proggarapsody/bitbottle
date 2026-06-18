// Package testhelpers — VCR recorder factory for integration tests.
// Uses gopkg.in/dnaeon/go-vcr.v2 to record real HTTP interactions once
// and replay them in CI without network access.
//
// Usage:
//
//	client, stop := testhelpers.NewVCRDoer(t, "path/to/cassette.yaml", "internal.host.invalid")
//	defer stop()
//	// client is an *http.Client that satisfies httpx.Doer
//	transport := httpx.New(client, baseURL, auth, errDecoder, ctPolicy, nil)
package testhelpers

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v2/cassette"
	"gopkg.in/dnaeon/go-vcr.v2/recorder"
)

// NewVCRDoer opens a cassette for replay and returns an *http.Client backed by
// the VCR recorder together with a stop function that must be deferred.
// *http.Client satisfies httpx.Doer (both expose Do(*http.Request)).
//
// Behaviour:
//   - Cassette absent + BITBOTTLE_RECORD != "1" → t.Skip (never fails CI).
//   - BITBOTTLE_RECORD=1 → record mode; cassette is written on stop().
//   - Cassette present → replay mode.
//
// Redaction is wired automatically in record mode: a SaveFilter applies Redact
// before the cassette YAML is written, stripping Authorization headers, Bearer
// tokens in bodies, and internalHost. Callers never have to remember to add the
// filter — the intended record path (make record-cassettes) is safe by default.
//
// internalHost is the real backend hostname (e.g. "git.example.invalid") that
// must be scrubbed from recorded URLs/bodies. Pass "" only when there is no
// host to redact (e.g. a public-host cassette).
//
// cassettePath must be the full path to the YAML file (the ".yaml" extension is
// stripped internally before being passed to go-vcr, which re-appends it).
func NewVCRDoer(t *testing.T, cassettePath, internalHost string) (*http.Client, func()) {
	t.Helper()

	// go-vcr appends ".yaml"; strip it so we don't end up with ".yaml.yaml".
	cassetteName := strings.TrimSuffix(cassettePath, ".yaml")

	record := os.Getenv("BITBOTTLE_RECORD") == "1"

	if !record {
		if _, err := os.Stat(cassettePath); os.IsNotExist(err) {
			t.Skip("cassette not recorded; run make record-cassettes")
			return nil, func() {}
		}
	}

	mode := recorder.ModeReplaying
	if record {
		mode = recorder.ModeRecording
	}

	rec, err := recorder.NewAsMode(cassetteName, mode, nil)
	if err != nil {
		t.Fatalf("vcr: open cassette %q: %v", cassettePath, err)
	}

	rec.SetMatcher(BodyAwareMatcher)
	rec.SkipRequestLatency = true

	// Safe-by-default: scrub secrets before anything is written to disk.
	rec.AddSaveFilter(func(i *cassette.Interaction) error {
		Redact(i, internalHost)
		return nil
	})

	client := &http.Client{Transport: rec}
	stop := func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("vcr: stop recorder: %v", err)
		}
	}

	return client, stop
}

// BodyAwareMatcher is a cassette.Matcher that compares the HTTP method, URL,
// and request body bytes. A mismatch on any of the three causes the interaction
// to be skipped, which surfaces as a "no matching interaction" error — proving
// that body-level changes (e.g. dropping the BQ-2 version field) are caught.
func BodyAwareMatcher(r *http.Request, i cassette.Request) bool {
	if r.Method != i.Method {
		return false
	}
	if r.URL.String() != i.URL {
		return false
	}
	if r.Body == nil || r.Body == http.NoBody {
		return i.Body == ""
	}
	actualBody, err := io.ReadAll(r.Body)
	if err != nil {
		return false
	}
	// Restore the body so the recorder/transport can read it again.
	r.Body = io.NopCloser(strings.NewReader(string(actualBody)))
	return string(actualBody) == i.Body
}

// Redact strips sensitive data from a cassette interaction before it is saved.
// It removes Authorization headers, Bearer/token values in bodies, and replaces
// internalHost with "REDACTED_HOST" in URLs and bodies.
//
// Wire it as a SaveFilter:
//
//	rec.AddSaveFilter(func(i *cassette.Interaction) error {
//	    testhelpers.Redact(i, "my-internal.host.invalid")
//	    return nil
//	})
func Redact(interaction *cassette.Interaction, internalHost string) {
	// Strip Authorization header from the recorded request.
	delete(interaction.Request.Headers, "Authorization")

	// Scrub bearer tokens from request and response bodies.
	interaction.Request.Body = redactTokens(interaction.Request.Body)
	interaction.Response.Body = redactTokens(interaction.Response.Body)

	// Replace internal hostname in URL and bodies.
	if internalHost != "" {
		interaction.URL = strings.ReplaceAll(
			interaction.URL, internalHost, "REDACTED_HOST",
		)
		interaction.Request.Body = strings.ReplaceAll(
			interaction.Request.Body, internalHost, "REDACTED_HOST",
		)
		interaction.Response.Body = strings.ReplaceAll(
			interaction.Response.Body, internalHost, "REDACTED_HOST",
		)
	}
}

// redactTokens removes "Bearer <value>" patterns from a string.
func redactTokens(s string) string {
	for {
		idx := strings.Index(s, "Bearer ")
		if idx < 0 {
			break
		}
		end := idx + len("Bearer ")
		for end < len(s) && s[end] != '"' && s[end] != '\'' && s[end] != ' ' && s[end] != '\n' {
			end++
		}
		s = s[:idx] + "REDACTED" + s[end:]
	}
	return s
}
