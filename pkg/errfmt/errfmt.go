// Package errfmt is the central renderer for user-facing CLI errors.
//
// Backend adapters classify wire errors into typed backend.DomainError values
// carrying a dotted ErrorCode. This package owns the mapping from code →
// human-readable title + hints, and writes them to the user via Render.
//
// Why a dedicated package: presentation concerns (TTY colour, line shape,
// hint phrasing) drift over time, and code-level translations live close
// to the data they describe. Keeping both in one place stops the cmd layer
// from duplicating ad-hoc message strings.
//
// Callers reach this package via cmdutil.ExplainError (legacy shim that
// most cmd/* packages still go through) or directly via Render in new
// code paths.
package errfmt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// entry holds the templated user-visible bits for one ErrorCode.
//
// Title may include {{.Host}} which is substituted from the DomainError.
// Hint lines may reference {{.Host}} the same way. Substitution is plain
// string replacement — keep templates simple, no Go templates needed.
type entry struct {
	title string
	hints []string
}

// catalogue is the source of truth for what each code looks like to the user.
// Add new codes alongside their backend.ErrorCode definition.
var catalogue = map[backend.ErrorCode]entry{
	backend.CodeAuthNoToken: {
		title: "No credentials configured for {{.Host}}.",
		hints: []string{"Run `bitbottle auth login --hostname {{.Host}}` to add a token."},
	},
	backend.CodeAuthInvalidToken: {
		title: "Authentication failed on {{.Host}}.",
		hints: []string{"Your token may be expired or revoked. Run `bitbottle auth login --hostname {{.Host}}` to refresh."},
	},
	backend.CodePermWriteRequired: {
		title: "Permission denied on {{.Host}}.",
		hints: []string{"Your token lacks write scope on this repository. Reissue it with the required permissions."},
	},
	backend.CodeRepoNotFound: {
		title: "Repository {{.ID}} not found on {{.Host}}.",
		hints: []string{"Check the slug casing. On Bitbucket Server use the project key, not the project name."},
	},
	backend.CodePRNotFound: {
		title: "Pull request #{{.ID}} not found on {{.Host}}.",
		hints: []string{"It may have been deleted. Run `bitbottle pr list` to see open PRs."},
	},
	backend.CodePRMergeConflict: {
		title: "Pull request #{{.ID}} cannot be merged: conflicts with the target branch on {{.Host}}.",
		hints: []string{"Resolve conflicts locally and push, then retry the merge."},
	},
	backend.CodePRMergeBehind: {
		title: "Pull request #{{.ID}} is behind its target branch on {{.Host}}.",
		hints: []string{"Update the source branch from base and push, then retry the merge."},
	},
	backend.CodePRCreateDuplicateBranch: {
		title: "A pull request for these branches already exists on {{.Host}}.",
		hints: []string{"Run `bitbottle pr list` to find it, or close it before creating a new one."},
	},
	backend.CodePRReviewerUnknown: {
		title: "One or more reviewers are not members of {{.Host}}.",
		hints: []string{"Check the slug spelling; users must already exist on the host."},
	},
	backend.CodePRAutoMergeBetaDisabled: {
		title: "Auto-merge is not enabled for this workspace on {{.Host}}.",
		hints: []string{"Ask your workspace admin to enable auto-merge in workspace settings."},
	},
	backend.CodeBranchProtected: {
		title: "Branch is protected on {{.Host}}.",
		hints: []string{"Ask an admin, or run `bitbottle branch protect list` to inspect the rules."},
	},
	backend.CodeHostUnsupported: {
		title: "`{{.Feature}}` is not available on {{.Host}}.",
		hints: []string{"This command targets a different Bitbucket flavour. Run `bitbottle config get backend_type` to verify the host."},
	},
	backend.CodeNetworkTLSUnknownAuthority: {
		title: "TLS verification failed for {{.Host}}.",
		hints: []string{"For self-signed CAs, pass `-k` or set `skip_tls_verify: true` in your host config."},
	},
	backend.CodeTransportTimeout: {
		title: "Request to {{.Host}} timed out.",
		hints: []string{"Network or VPN may be slow or down. Retry; pass `--debug` for transport details."},
	},
	backend.CodeInvalidRequest: {
		title: "The request was rejected by {{.Host}} as invalid.",
		hints: []string{"Check that all required fields are set and values are within the expected range."},
	},
}

// Render writes a friendly error explanation to ios.ErrOut.
//
// Resolution order:
//  1. nil → no-op (safe to call from generic shims).
//  2. DomainError whose Code is present in the catalogue → templated
//     title/cause/hint output.
//  3. DomainError whose Code is unset OR not in the catalogue, but whose
//     Kind matches one of backend.ErrConflict / ErrUnsupportedOnHost /
//     ErrPermission / ErrNotFound → Kind-based fallback line.
//  4. anything else (including DomainErrors with no Code and no
//     recognised Kind, or non-DomainError values) → raw error string.
//
// Every server- or config-derived field (Message, Host, Resource, ID,
// Feature) is passed through sanitise before reaching ios.ErrOut to
// neutralise terminal escape injection (CWE-150 / CWE-117).
func Render(ios *iostreams.IOStreams, err error) {
	if err == nil {
		return
	}
	var de *backend.DomainError
	if errors.As(err, &de) {
		if e, ok := catalogue[de.Code]; ok {
			fmt.Fprintln(ios.ErrOut, expand(e.title, de))
			if cause := de.Message; cause != "" {
				fmt.Fprintln(ios.ErrOut, "Cause: "+sanitise(cause))
			}
			for _, h := range e.hints {
				fmt.Fprintln(ios.ErrOut, "Hint:  "+expand(h, de))
			}
			return
		}
		if renderByKind(ios, de) {
			return
		}
	}
	fmt.Fprintln(ios.ErrOut, sanitise(err.Error()))
}

// renderByKind is the pre-catalogue fallback. It mirrors the old
// cmdutil.ExplainError behaviour so DomainErrors that haven't been
// stamped with a Code yet still produce a humanised line. Returns true
// when it handled the error.
//
// New code should attach an ErrorCode at the classification layer; this
// fallback exists only to keep migrations one-PR-at-a-time.
func renderByKind(ios *iostreams.IOStreams, de *backend.DomainError) bool {
	switch {
	case errors.Is(de, backend.ErrConflict):
		fmt.Fprintln(ios.ErrOut, "A resource with these parameters already exists.")
	case errors.Is(de, backend.ErrUnsupportedOnHost):
		feature := "this feature"
		if de.Feature != "" {
			feature = sanitise(de.Feature)
		}
		host := ""
		if de.Host != "" {
			host = " on " + sanitise(de.Host)
		}
		fmt.Fprintf(ios.ErrOut, "%s is not available%s.\n", feature, host)
	case errors.Is(de, backend.ErrPermission):
		if de.Host != "" {
			fmt.Fprintf(ios.ErrOut, "Permission denied on %s. Check your token scopes or re-run `bitbottle auth login`.\n", sanitise(de.Host))
		} else {
			fmt.Fprintln(ios.ErrOut, "Permission denied. Check your token scopes or re-run `bitbottle auth login`.")
		}
	case errors.Is(de, backend.ErrNotFound):
		if de.Resource != "" && de.ID != "" {
			fmt.Fprintf(ios.ErrOut, "%s %q not found.\n", sanitise(de.Resource), sanitise(de.ID))
		} else {
			fmt.Fprintln(ios.ErrOut, "Resource not found.")
		}
	default:
		return false
	}
	return true
}

// HintsFor returns the catalogue's hint strings for de.Code with
// template placeholders ({{.Host}}, {{.ID}}, {{.Resource}}, {{.Feature}})
// expanded against the DomainError. Returns nil when de is nil, when the
// Code is empty, or when no catalogue entry exists for the code.
//
// MCP and other structured-output surfaces use this to populate the
// "hints" field of their error envelope so clients see the same
// remediation steps as CLI users without bundling their own catalogue.
func HintsFor(de *backend.DomainError) []string {
	if de == nil || de.Code == "" {
		return nil
	}
	e, ok := catalogue[de.Code]
	if !ok || len(e.hints) == 0 {
		return nil
	}
	out := make([]string, 0, len(e.hints))
	for _, h := range e.hints {
		out = append(out, expand(h, de))
	}
	return out
}

// expand substitutes template placeholders with values from de.
// Supported placeholders: {{.Host}}, {{.ID}}, {{.Resource}}, {{.Feature}}.
// Add new placeholders here before referencing them in catalogue entries.
//
// Every substituted value is sanitised because it may originate from
// untrusted sources: Host comes from local config / git remote URLs /
// the -R flag (could carry escapes from a hostile .git/config), and
// ID/Resource/Feature can flow from server response bodies (a hostile
// or compromised Bitbucket Server could embed CSI / OSC sequences).
// See sanitise for the threat model (CWE-150 / CWE-117).
func expand(tmpl string, de *backend.DomainError) string {
	r := strings.NewReplacer(
		"{{.Host}}", sanitise(de.Host),
		"{{.ID}}", sanitise(de.ID),
		"{{.Resource}}", sanitise(de.Resource),
		"{{.Feature}}", sanitise(de.Feature),
	)
	return r.Replace(tmpl)
}

// sanitise strips bytes that would let a hostile server or local config
// inject terminal escape sequences into stderr (CWE-150 / CWE-117).
// Removed: C0 controls (0x00–0x1F) except \t and \n, the DEL byte (0x7F),
// and C1 controls (0x80–0x9F). \t and \n survive because they're load-
// bearing in legitimate output (line splitting, indentation in upstream
// messages). The lone ESC byte (0x1B) is part of C0 and is therefore
// removed, which neutralises CSI / OSC 8 / OSC 52 / bracketed-paste
// sequences in one pass.
func sanitise(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t' || r == '\n':
			b.WriteRune(r)
		case r < 0x20 || r == 0x7F:
			// C0 control + DEL — drop
		case r >= 0x80 && r <= 0x9F:
			// C1 control — drop
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
