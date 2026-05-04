package cmdutil

import (
	"errors"
	"fmt"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// ExplainError writes a user-actionable message to ios.ErrOut for known
// DomainError classes, falling back to the raw error for anything else.
//
// Why this exists: the API layer already classifies wire errors into typed
// DomainErrors (ErrConflict, ErrNotFound, etc.), but historically only the
// MCP surface consumed that information. CLI users saw raw "HTTP 409" strings.
// ExplainError is the presentation layer that finally collects the leverage
// the typed-error system was designed for.
func ExplainError(ios *iostreams.IOStreams, err error) {
	if err == nil {
		return
	}

	// Pull structured fields when we have them so messages can name what
	// failed instead of just how.
	var de *backend.DomainError
	_ = errors.As(err, &de)

	switch {
	case errors.Is(err, backend.ErrConflict):
		fmt.Fprintln(ios.ErrOut, "A resource with these parameters already exists.")
	case errors.Is(err, backend.ErrUnsupportedOnHost):
		feature := "this feature"
		if de != nil && de.Feature != "" {
			feature = de.Feature
		}
		host := ""
		if de != nil && de.Host != "" {
			host = " on " + de.Host
		}
		fmt.Fprintf(ios.ErrOut, "%s is not available%s.\n", feature, host)
	case errors.Is(err, backend.ErrPermission):
		if de != nil && de.Host != "" {
			fmt.Fprintf(ios.ErrOut, "Permission denied on %s. Check your token scopes or re-run `bitbottle auth login`.\n", de.Host)
			return
		}
		fmt.Fprintln(ios.ErrOut, "Permission denied. Check your token scopes or re-run `bitbottle auth login`.")
	case errors.Is(err, backend.ErrNotFound):
		if de != nil && de.Resource != "" && de.ID != "" {
			fmt.Fprintf(ios.ErrOut, "%s %q not found.\n", de.Resource, de.ID)
			return
		}
		fmt.Fprintln(ios.ErrOut, "Resource not found.")
	default:
		fmt.Fprintln(ios.ErrOut, err)
	}
}
