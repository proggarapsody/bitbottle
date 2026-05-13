package backend_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// ── NormaliseEmoji ───────────────────────────────────────────────────────────

func TestNormaliseEmoji(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		{":thumbsup:", "thumbs_up"},
		{"thumbsup", "thumbs_up"},
		{"thumbs_up", "thumbs_up"},
		{":thumbsdown:", "thumbs_down"},
		{"thumbsdown", "thumbs_down"},
		{"thumbs_down", "thumbs_down"},
		{":heart:", "heart"},
		{"heart", "heart"},
		{":smile:", "laugh"},
		{"smile", "laugh"},
		{"laugh", "laugh"},
		{":tada:", "hooray"},
		{"tada", "hooray"},
		{"hooray", "hooray"},
		{":confused:", "confused"},
		{"confused", "confused"},
		// unknown passthrough
		{"unknown_emoji", "unknown_emoji"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, backend.NormaliseEmoji(tc.input))
		})
	}
}
