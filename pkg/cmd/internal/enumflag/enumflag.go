// Package enumflag provides a pflag.Value implementation that restricts a
// string flag to a fixed set of allowed values, rejecting off-enum input
// (including the empty string) at flag-parse time rather than silently
// accepting it downstream.
package enumflag

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// enumValue is a pflag.Value that accepts only one of a fixed set of strings.
type enumValue struct {
	allowed         []string
	target          *string
	caseInsensitive bool
}

// New returns a pflag.Value that writes through to target and rejects any
// value not in allowed. When caseInsensitive is true, comparison is
// case-folded but the original (caller-typed) value is stored in target so
// downstream normalisation (e.g. mapPRState) sees exactly what the user
// passed. The empty string is never a member of allowed unless explicitly
// listed, so "" is rejected like any other off-enum value.
func New(allowed []string, target *string, caseInsensitive bool) pflag.Value {
	return &enumValue{allowed: allowed, target: target, caseInsensitive: caseInsensitive}
}

// String returns the current value.
func (e *enumValue) String() string {
	if e.target == nil {
		return ""
	}
	return *e.target
}

// Set validates v against the allowed set and stores it, or returns a clear
// error listing the permitted values.
func (e *enumValue) Set(v string) error {
	for _, a := range e.allowed {
		if v == a || (e.caseInsensitive && strings.EqualFold(v, a)) {
			*e.target = v
			return nil
		}
	}
	return fmt.Errorf("invalid value %q: must be one of %s", v, strings.Join(e.allowed, ", "))
}

// Type reports the flag type name shown in usage output.
func (e *enumValue) Type() string {
	return "string"
}
