// Package argval holds typed argument extractors for MCP tool handlers.
//
// Before argval, handlers reached into the raw argument map with
// req.GetInt / req.GetString and ad-hoc presence checks. Those checks
// could not distinguish "argument missing" from "argument present but
// wrong type" (both collapsed to the Go zero value), so a string passed
// for a numeric id reported "missing required parameter: id", and an
// explicit id:0 looked missing too. argval reads the raw map directly
// and returns a structured *Error carrying a dotted code, the field
// name, and the offending value, so MCP clients can branch on
// {code, field, got} instead of parsing prose.
package argval

import (
	"fmt"
	"strings"
)

// Error is a client-side argument-validation failure. It maps to the MCP
// tool-error envelope as {code, field, got, message}. Code is a dotted
// token (e.g. "arg.invalid_type") so clients can branch without parsing
// the human-readable Message.
type Error struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Got     string `json:"got,omitempty"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

// Error codes. These are client-side validation codes, distinct from the
// backend.ErrorCode catalogue (which classifies server responses).
const (
	CodeMissing      = "arg.missing"
	CodeInvalidType  = "arg.invalid_type"
	CodeOutOfRange   = "arg.out_of_range"
	CodeInvalidValue = "arg.invalid_value"
)

func missing(field string) *Error {
	return &Error{Code: CodeMissing, Field: field, Message: fmt.Sprintf("missing required parameter: %s", field)}
}

func invalidType(field, got, want string) *Error {
	return &Error{
		Code:    CodeInvalidType,
		Field:   field,
		Got:     got,
		Message: fmt.Sprintf("%s must be %s", field, want),
	}
}

func outOfRange(field, got, constraint string) *Error {
	return &Error{
		Code:    CodeOutOfRange,
		Field:   field,
		Got:     got,
		Message: fmt.Sprintf("%s %s", field, constraint),
	}
}

func invalidValue(field, got, why string) *Error {
	return &Error{
		Code:    CodeInvalidValue,
		Field:   field,
		Got:     got,
		Message: fmt.Sprintf("%s %s", field, why),
	}
}

// IntOption configures Int.
type IntOption func(*intOpts)

type intOpts struct {
	required bool
	hasMin   bool
	min      int
	hasMax   bool
	max      int
}

// Required marks an Int argument as required. Without it, a missing key
// yields (0, false, nil) so callers can apply their own default.
func Required() IntOption { return func(o *intOpts) { o.required = true } }

// Min sets an inclusive lower bound for an Int argument.
func Min(n int) IntOption { return func(o *intOpts) { o.hasMin = true; o.min = n } }

// Max sets an inclusive upper bound for an Int argument.
func Max(n int) IntOption { return func(o *intOpts) { o.hasMax = true; o.max = n } }

// Int extracts an integer argument from a raw argument map.
//
// It distinguishes three failure modes that the old req.GetInt collapsed:
//
//   - missing key       → present=false; *Error{Code: arg.missing} only when Required().
//   - present non-number → *Error{Code: arg.invalid_type, got: "<type>"}.
//   - out of [Min,Max]   → *Error{Code: arg.out_of_range}.
//
// An explicit 0 (or any in-range value) is reported present=true, fixing
// the Go zero-value "falsely missing" bug. JSON numbers arrive as
// float64; a non-integral float64 (e.g. 1.5) is rejected as invalid_type.
func Int(args map[string]any, field string, opts ...IntOption) (value int, present bool, err *Error) {
	var o intOpts
	for _, opt := range opts {
		opt(&o)
	}

	raw, ok := args[field]
	if !ok || raw == nil {
		if o.required {
			return 0, false, missing(field)
		}
		return 0, false, nil
	}

	n, convErr := toInt(raw)
	if convErr != nil {
		return 0, false, invalidType(field, typeName(raw), "integer")
	}

	if o.hasMin && n < o.min {
		return 0, true, outOfRange(field, fmt.Sprintf("%d", n), fmt.Sprintf("must be >= %d", o.min))
	}
	if o.hasMax && n > o.max {
		return 0, true, outOfRange(field, fmt.Sprintf("%d", n), fmt.Sprintf("must be <= %d", o.max))
	}
	return n, true, nil
}

// toInt converts a JSON-decoded value to an int, rejecting non-integral
// floats and non-numeric types. Numeric strings are NOT accepted: an MCP
// numeric arg arrives as a JSON number (float64), so a string in a numeric
// slot is a genuine type error the caller wants surfaced.
func toInt(raw any) (int, error) {
	switch v := raw.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		if v != float64(int(v)) {
			return 0, fmt.Errorf("non-integral")
		}
		return int(v), nil
	case float32:
		f := float64(v)
		if f != float64(int(f)) {
			return 0, fmt.Errorf("non-integral")
		}
		return int(f), nil
	default:
		return 0, fmt.Errorf("not a number")
	}
}

func typeName(raw any) string {
	switch raw.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64, float32, int, int64:
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", raw)
	}
}

// Hash extracts and validates a commit-hash argument. The value must be
// present, non-empty, at least minLen characters, and composed solely of
// hex digits. minLen defaults to 7 when <= 0 (the conventional short-hash
// floor). Rejecting "a" and "NOT_HEX" client-side stops them reaching the
// Cloud API as a generic 404.
func Hash(args map[string]any, field string, minLen int) (string, *Error) {
	if minLen <= 0 {
		minLen = 7
	}
	raw, ok := args[field]
	if !ok || raw == nil {
		return "", missing(field)
	}
	s, ok := raw.(string)
	if !ok {
		return "", invalidType(field, typeName(raw), "a string")
	}
	if s == "" {
		return "", missing(field)
	}
	if len(s) < minLen {
		return "", invalidValue(field, s, fmt.Sprintf("must be a commit hash of at least %d hex characters", minLen))
	}
	for _, r := range s {
		if !isHex(r) {
			return "", invalidValue(field, s, "must be a hexadecimal commit hash")
		}
	}
	return s, nil
}

func isHex(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// RefName extracts and validates a Git branch/ref name. It rejects names
// that Git refuses by spec (git check-ref-format rules, the subset that
// matters for branch creation):
//
//   - empty, or "/" / leading / trailing / doubled slash
//   - "." at the start of any path component, or a component ending ".lock"
//   - the sequences "..", "@{", a lone "@"
//   - ASCII control chars and the set space ~ ^ : ? * [ \
//
// This stops "/" and "feature/" reaching the API as a generic 404.
func RefName(args map[string]any, field string) (string, *Error) {
	raw, ok := args[field]
	if !ok || raw == nil {
		return "", missing(field)
	}
	s, ok := raw.(string)
	if !ok {
		return "", invalidType(field, typeName(raw), "a string")
	}
	if s == "" {
		return "", missing(field)
	}
	if why := invalidRefReason(s); why != "" {
		return "", invalidValue(field, s, why)
	}
	return s, nil
}

func invalidRefReason(s string) string {
	if s == "@" {
		return "is not a valid ref name"
	}
	if strings.HasPrefix(s, "/") || strings.HasSuffix(s, "/") {
		return "must not start or end with a slash"
	}
	if strings.Contains(s, "//") {
		return "must not contain consecutive slashes"
	}
	if strings.Contains(s, "..") || strings.Contains(s, "@{") {
		return "must not contain '..' or '@{'"
	}
	if strings.HasSuffix(s, ".") || strings.HasSuffix(s, ".lock") {
		return "must not end with '.' or '.lock'"
	}
	for _, r := range s {
		switch r {
		case ' ', '~', '^', ':', '?', '*', '[', '\\':
			return fmt.Sprintf("must not contain %q", string(r))
		}
		if r < 0x20 || r == 0x7f {
			return "must not contain control characters"
		}
	}
	for _, comp := range strings.Split(s, "/") {
		if comp == "" {
			return "must not contain empty path components"
		}
		if strings.HasPrefix(comp, ".") {
			return "path components must not start with '.'"
		}
		if strings.HasSuffix(comp, ".lock") {
			return "path components must not end with '.lock'"
		}
	}
	return ""
}

// EnumOneOf extracts a string argument that must be one of allowed. An
// empty allowed set, or an empty string member, is a programming error:
// EnumOneOf panics so the "" member that produced merge_pr's
// "must be one of , merge, squash" message can never be reintroduced.
// A missing key returns ("", nil) when not required so callers can apply
// a default; pass Required-style presence checking separately if needed.
func EnumOneOf(args map[string]any, field string, allowed []string) (string, *Error) {
	if len(allowed) == 0 {
		panic("argval.EnumOneOf: empty allowed set")
	}
	for _, a := range allowed {
		if a == "" {
			panic("argval.EnumOneOf: empty string is not a valid enum member; treat empty as 'use default' before calling")
		}
	}
	raw, ok := args[field]
	if !ok || raw == nil {
		return "", nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", invalidType(field, typeName(raw), "a string")
	}
	if s == "" {
		return "", nil
	}
	for _, a := range allowed {
		if s == a {
			return s, nil
		}
	}
	return "", invalidValue(field, s, "must be one of "+strings.Join(allowed, ", "))
}

// MutuallyRequired enforces that two arguments are supplied together: if
// either field or dependency is present (non-empty / non-nil), the other
// must be too. This makes add_pr_comment's inline-anchor check symmetric —
// inline_line without inline_path is now caught, not just the reverse.
func MutuallyRequired(args map[string]any, fieldA, fieldB string) *Error {
	a := isPresent(args, fieldA)
	b := isPresent(args, fieldB)
	switch {
	case a && !b:
		return &Error{Code: CodeMissing, Field: fieldB, Message: fmt.Sprintf("%s requires %s", fieldA, fieldB)}
	case b && !a:
		return &Error{Code: CodeMissing, Field: fieldA, Message: fmt.Sprintf("%s requires %s", fieldB, fieldA)}
	default:
		return nil
	}
}

// OneOfRequired enforces that at least one of fields is present. Used by
// update_pr so a call with neither title nor body is rejected client-side
// ("nothing to update") instead of hitting the API.
func OneOfRequired(args map[string]any, fields ...string) *Error {
	for _, f := range fields {
		if isPresent(args, f) {
			return nil
		}
	}
	return &Error{
		Code:    CodeMissing,
		Field:   strings.Join(fields, "|"),
		Message: "nothing to update: provide at least one of " + strings.Join(fields, ", "),
	}
}

// isPresent reports whether a key carries a meaningful value: present, not
// nil, and (for strings) non-empty. A numeric 0 counts as present.
func isPresent(args map[string]any, field string) bool {
	raw, ok := args[field]
	if !ok || raw == nil {
		return false
	}
	if s, ok := raw.(string); ok {
		return s != ""
	}
	return true
}
