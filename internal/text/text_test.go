package text_test

import (
	"testing"
	"time"

	"github.com/aleksey/bitbottle/internal/text"
	"github.com/stretchr/testify/assert"
)

func TestTruncate_Short(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "hi", text.Truncate("hi", 10))
}

func TestTruncate_Exact(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "hello", text.Truncate("hello", 5))
}

func TestTruncate_Long(t *testing.T) {
	t.Parallel()
	got := text.Truncate("hello world", 5)
	assert.Equal(t, "hell…", got)
}

func TestTruncate_Empty(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", text.Truncate("", 5))
}

func TestPluralize_One(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "item", text.Pluralize(1, "item", "items"))
}

func TestPluralize_Many(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "items", text.Pluralize(5, "item", "items"))
}

func TestRelativeTime_JustNow(t *testing.T) {
	t.Parallel()
	got := text.RelativeTime(time.Now().Add(-30 * time.Second))
	assert.Equal(t, "just now", got)
}

func TestRelativeTime_Hours(t *testing.T) {
	t.Parallel()
	got := text.RelativeTime(time.Now().Add(-3 * time.Hour))
	assert.Contains(t, got, "3 hours")
}

func TestRelativeTime_Days(t *testing.T) {
	t.Parallel()
	got := text.RelativeTime(time.Now().Add(-48 * time.Hour))
	assert.Contains(t, got, "2 days")
}

func TestPadRight_Shorter(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "hi   ", text.PadRight("hi", 5))
}

func TestPadRight_Exact(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "hello", text.PadRight("hello", 5))
}

func TestPadRight_Longer(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "toolong", text.PadRight("toolong", 3))
}
