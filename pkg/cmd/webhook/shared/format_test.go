package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/pkg/cmd/webhook/shared"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

func TestParseEvents_Empty(t *testing.T) {
	t.Parallel()
	assert.Nil(t, shared.ParseEvents(""))
}

func TestParseEvents_TrimsAndDedupes(t *testing.T) {
	t.Parallel()
	got := shared.ParseEvents(" repo:push , pullrequest:created ,repo:push,, pullrequest:created ")
	assert.Equal(t, []string{"repo:push", "pullrequest:created"}, got)
}

func TestParseEvents_AllWhitespaceReturnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, shared.ParseEvents("   ,  ,  "))
}

func TestParseEvents_PreservesOrder(t *testing.T) {
	t.Parallel()
	got := shared.ParseEvents("c,a,b")
	assert.Equal(t, []string{"c", "a", "b"}, got)
}

func TestWebhookActiveColor_TTY(t *testing.T) {
	t.Parallel()
	colorize := shared.WebhookActiveColor(iostreams.TestTTY())

	cases := map[string]string{
		"true":  "\033[32mtrue\033[0m",
		"false": "\033[31mfalse\033[0m",
		// Edge cases:
		"":     "",
		"True": "True", // ColorFunc receives fmt.Sprintf("%v", bool); never "True"
		"yes":  "yes",  // unknown value passes through
	}
	for active, want := range cases {
		assert.Equal(t, want, colorize(active), "active=%q", active)
	}
}

func TestWebhookActiveColor_NonTTY(t *testing.T) {
	t.Parallel()
	colorize := shared.WebhookActiveColor(iostreams.Test())
	assert.Equal(t, "true", colorize("true"))
	assert.Equal(t, "false", colorize("false"))
}
