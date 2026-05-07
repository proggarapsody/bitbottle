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
}

// Render writes a friendly error explanation to ios.ErrOut.
//
// Resolution order:
//  1. nil → no-op (safe to call from generic shims).
//  2. DomainError with a code in the catalogue → templated title/cause/hint.
//  3. DomainError with a recognised Kind but no code → legacy Kind-based
//     fallback so commands not yet migrated still get a friendly line.
//  4. anything else → raw error string.
func Render(ios *iostreams.IOStreams, err error) {
	if err == nil {
		return
	}
	var de *backend.DomainError
	if errors.As(err, &de) {
		if e, ok := catalogue[de.Code]; ok {
			fmt.Fprintln(ios.ErrOut, expand(e.title, de))
			if cause := de.Message; cause != "" {
				fmt.Fprintln(ios.ErrOut, "Cause: "+cause)
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
	fmt.Fprintln(ios.ErrOut, err)
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
			feature = de.Feature
		}
		host := ""
		if de.Host != "" {
			host = " on " + de.Host
		}
		fmt.Fprintf(ios.ErrOut, "%s is not available%s.\n", feature, host)
	case errors.Is(de, backend.ErrPermission):
		if de.Host != "" {
			fmt.Fprintf(ios.ErrOut, "Permission denied on %s. Check your token scopes or re-run `bitbottle auth login`.\n", de.Host)
			return true
		}
		fmt.Fprintln(ios.ErrOut, "Permission denied. Check your token scopes or re-run `bitbottle auth login`.")
	case errors.Is(de, backend.ErrNotFound):
		if de.Resource != "" && de.ID != "" {
			fmt.Fprintf(ios.ErrOut, "%s %q not found.\n", de.Resource, de.ID)
			return true
		}
		fmt.Fprintln(ios.ErrOut, "Resource not found.")
	default:
		return false
	}
	return true
}

// expand substitutes {{.Host}} with the DomainError host. Kept as a tiny
// helper so callers can read the template strings literally above.
func expand(tmpl string, de *backend.DomainError) string {
	return strings.ReplaceAll(tmpl, "{{.Host}}", de.Host)
}
