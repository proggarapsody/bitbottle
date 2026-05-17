package text_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/internal/text"
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

