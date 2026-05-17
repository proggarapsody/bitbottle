package cmdutil

import (
	"fmt"
	"io"
)

// PartialWarn writes a warning to w when items were collected but pagination
// ended with an error. Call it after rendering and before returning listErr.
// It is a no-op when listErr is nil or n == 0.
func PartialWarn(w io.Writer, n int, listErr error) {
	if listErr != nil && n > 0 {
		fmt.Fprintf(w, "warning: partial results (%d items) — %s\n", n, listErr)
	}
}
