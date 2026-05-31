package argval

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInt_Missing(t *testing.T) {
	t.Parallel()
	// not required → no error, present=false
	v, present, err := Int(map[string]any{}, "id")
	require.Nil(t, err)
	assert.False(t, present)
	assert.Equal(t, 0, v)

	// required → arg.missing
	_, present, err = Int(map[string]any{}, "id", Required())
	require.NotNil(t, err)
	assert.False(t, present)
	assert.Equal(t, CodeMissing, err.Code)
	assert.Equal(t, "id", err.Field)
}

func TestInt_WrongType_String(t *testing.T) {
	t.Parallel()
	// MCP-06: a string in a numeric slot is a type error, not "missing".
	_, _, err := Int(map[string]any{"id": "abc"}, "id", Required())
	require.NotNil(t, err)
	assert.Equal(t, CodeInvalidType, err.Code)
	assert.Equal(t, "id", err.Field)
	assert.Equal(t, "string", err.Got)
	assert.Contains(t, err.Message, "id must be integer")
}

func TestInt_NonIntegralFloat(t *testing.T) {
	t.Parallel()
	_, _, err := Int(map[string]any{"id": float64(1.5)}, "id", Required())
	require.NotNil(t, err)
	assert.Equal(t, CodeInvalidType, err.Code)
}

func TestInt_ZeroIsPresent(t *testing.T) {
	t.Parallel()
	// MCP-07: explicit 0 must be accepted as present.
	v, present, err := Int(map[string]any{"id": float64(0)}, "id", Required())
	require.Nil(t, err)
	assert.True(t, present)
	assert.Equal(t, 0, v)
}

func TestInt_Negative_RejectedByMin(t *testing.T) {
	t.Parallel()
	// MCP-08: negative id rejected client-side when Min(1) is set.
	_, _, err := Int(map[string]any{"id": float64(-3)}, "id", Required(), Min(1))
	require.NotNil(t, err)
	assert.Equal(t, CodeOutOfRange, err.Code)
	assert.Equal(t, "id", err.Field)
	assert.Equal(t, "-3", err.Got)
}

func TestInt_RespectsMinMax(t *testing.T) {
	t.Parallel()
	v, present, err := Int(map[string]any{"limit": float64(50)}, "limit", Min(1), Max(100))
	require.Nil(t, err)
	assert.True(t, present)
	assert.Equal(t, 50, v)

	_, _, err = Int(map[string]any{"limit": float64(101)}, "limit", Min(1), Max(100))
	require.NotNil(t, err)
	assert.Equal(t, CodeOutOfRange, err.Code)
}

func TestInt_AcceptsNativeInt(t *testing.T) {
	t.Parallel()
	v, present, err := Int(map[string]any{"id": 42}, "id", Required())
	require.Nil(t, err)
	assert.True(t, present)
	assert.Equal(t, 42, v)
}

func TestHash_Valid(t *testing.T) {
	t.Parallel()
	v, err := Hash(map[string]any{"hash": "abc1234"}, "hash", 7)
	require.Nil(t, err)
	assert.Equal(t, "abc1234", v)
}

func TestHash_TooShort(t *testing.T) {
	t.Parallel()
	// MCP-11: "a" is too short.
	_, err := Hash(map[string]any{"hash": "a"}, "hash", 7)
	require.NotNil(t, err)
	assert.Equal(t, CodeInvalidValue, err.Code)
	assert.Equal(t, "hash", err.Field)
}

func TestHash_NotHex(t *testing.T) {
	t.Parallel()
	// MCP-11: "NOT_HEX" is not hexadecimal.
	_, err := Hash(map[string]any{"hash": "NOT_HEX"}, "hash", 7)
	require.NotNil(t, err)
	assert.Equal(t, CodeInvalidValue, err.Code)
}

func TestHash_Missing(t *testing.T) {
	t.Parallel()
	_, err := Hash(map[string]any{}, "hash", 7)
	require.NotNil(t, err)
	assert.Equal(t, CodeMissing, err.Code)
}

func TestHash_DefaultMinLen(t *testing.T) {
	t.Parallel()
	// minLen <= 0 defaults to 7.
	_, err := Hash(map[string]any{"hash": "abc12"}, "hash", 0)
	require.NotNil(t, err)
	assert.Equal(t, CodeInvalidValue, err.Code)
}

func TestRefName_Valid(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"main", "feature/new", "release/1.2.3"} {
		v, err := RefName(map[string]any{"name": name}, "name")
		require.Nilf(t, err, "expected %q to be valid", name)
		assert.Equal(t, name, v)
	}
}

func TestRefName_Invalid(t *testing.T) {
	t.Parallel()
	// MCP-12: slashes and Git-invalid refs rejected.
	for _, name := range []string{
		"/", "/leading", "trailing/", "a//b", "a..b", "feat~1",
		"feat space", ".hidden", "x.lock", "feat^", "feat:bad", "a@{0}",
	} {
		_, err := RefName(map[string]any{"name": name}, "name")
		require.NotNilf(t, err, "expected %q to be rejected", name)
		assert.Equal(t, CodeInvalidValue, err.Code, "name=%q", name)
		assert.Equal(t, "name", err.Field)
	}
}

func TestRefName_Missing(t *testing.T) {
	t.Parallel()
	_, err := RefName(map[string]any{}, "name")
	require.NotNil(t, err)
	assert.Equal(t, CodeMissing, err.Code)
}

func TestEnumOneOf_Valid(t *testing.T) {
	t.Parallel()
	v, err := EnumOneOf(map[string]any{"strategy": "squash"}, "strategy", []string{"merge", "squash", "rebase"})
	require.Nil(t, err)
	assert.Equal(t, "squash", v)
}

func TestEnumOneOf_EmptyTreatedAsDefault(t *testing.T) {
	t.Parallel()
	// MCP-09: empty/missing strategy returns "" with no error (use default).
	v, err := EnumOneOf(map[string]any{"strategy": ""}, "strategy", []string{"merge", "squash", "rebase"})
	require.Nil(t, err)
	assert.Equal(t, "", v)

	v, err = EnumOneOf(map[string]any{}, "strategy", []string{"merge", "squash", "rebase"})
	require.Nil(t, err)
	assert.Equal(t, "", v)
}

func TestEnumOneOf_Invalid_NoBareComma(t *testing.T) {
	t.Parallel()
	// MCP-09: error message lists only real members — no leading bare comma.
	_, err := EnumOneOf(map[string]any{"strategy": "bogus"}, "strategy", []string{"merge", "squash", "rebase"})
	require.NotNil(t, err)
	assert.Equal(t, CodeInvalidValue, err.Code)
	assert.Contains(t, err.Message, "must be one of merge, squash, rebase")
	assert.NotContains(t, err.Message, "one of ,")
}

func TestEnumOneOf_PanicsOnEmptyMember(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() {
		_, _ = EnumOneOf(map[string]any{"strategy": "x"}, "strategy", []string{"", "merge"})
	})
}

func TestMutuallyRequired(t *testing.T) {
	t.Parallel()
	// both present → ok
	require.Nil(t, MutuallyRequired(map[string]any{"inline_path": "f.go", "inline_line": float64(3)}, "inline_path", "inline_line"))
	// neither → ok
	require.Nil(t, MutuallyRequired(map[string]any{}, "inline_path", "inline_line"))

	// MCP-10: inline_line without inline_path → error naming inline_path
	err := MutuallyRequired(map[string]any{"inline_line": float64(3)}, "inline_path", "inline_line")
	require.NotNil(t, err)
	assert.Equal(t, "inline_path", err.Field)

	// reverse: inline_path without inline_line → error naming inline_line
	err = MutuallyRequired(map[string]any{"inline_path": "f.go"}, "inline_path", "inline_line")
	require.NotNil(t, err)
	assert.Equal(t, "inline_line", err.Field)
}

func TestOneOfRequired(t *testing.T) {
	t.Parallel()
	// MCP-13: neither title nor body → error
	err := OneOfRequired(map[string]any{}, "title", "body")
	require.NotNil(t, err)
	assert.Equal(t, CodeMissing, err.Code)
	assert.Contains(t, err.Message, "nothing to update")

	// title present → ok
	require.Nil(t, OneOfRequired(map[string]any{"title": "x"}, "title", "body"))
	// body present → ok
	require.Nil(t, OneOfRequired(map[string]any{"body": "x"}, "title", "body"))
	// empty strings count as absent
	require.NotNil(t, OneOfRequired(map[string]any{"title": "", "body": ""}, "title", "body"))
}
