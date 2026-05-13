package backend

// emojiToCanonical maps Bitbucket Server emoticon values (colon-wrapped or
// otherwise) to the canonical underscore shortcodes used everywhere in the
// domain layer.
var emojiToCanonical = map[string]string{
	":thumbsup:":   "thumbs_up",
	"thumbsup":     "thumbs_up",
	"thumbs_up":    "thumbs_up",
	":thumbsdown:": "thumbs_down",
	"thumbsdown":   "thumbs_down",
	"thumbs_down":  "thumbs_down",
	":heart:":      "heart",
	"heart":        "heart",
	":smile:":      "laugh",
	"smile":        "laugh",
	"laugh":        "laugh",
	":tada:":       "hooray",
	"tada":         "hooray",
	"hooray":       "hooray",
	":confused:":   "confused",
	"confused":     "confused",
}

// NormaliseEmoji converts any accepted emoji form (colon-wrapped Server
// values, shortcodes with colons, or canonical underscore form) to the
// canonical underscore form. Returns the input unchanged when the emoji is
// not in the known set.
func NormaliseEmoji(emoji string) string {
	if c, ok := emojiToCanonical[emoji]; ok {
		return c
	}
	// Also try stripping leading/trailing colons for colon-form input.
	stripped := emoji
	if len(stripped) > 1 && stripped[0] == ':' && stripped[len(stripped)-1] == ':' {
		inner := stripped[1 : len(stripped)-1]
		if c, ok := emojiToCanonical[inner]; ok {
			return c
		}
		if c, ok := emojiToCanonical[":"+inner+":"]; ok {
			return c
		}
	}
	return emoji
}
