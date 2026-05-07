// Package shared holds helpers used across webhook subcommands.
package shared

import (
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// WebhookActiveColor maps a webhook's Active boolean (rendered as "true" or
// "false" by fmt.Sprintf("%v", ...)) to a color. Active webhooks are green
// (live, will fire) and inactive ones red (silent until re-enabled). This
// helper takes the already-formatted string because that's what
// format.Field.ColorFunc receives.
func WebhookActiveColor(ios *iostreams.IOStreams) func(string) string {
	return func(active string) string {
		switch active {
		case "true":
			return ios.ColorGreen(active)
		case "false":
			return ios.ColorRed(active)
		default:
			return active
		}
	}
}

// WebhookFields constructs the formatter shared by `webhook list` and
// `webhook view` so JSON field names and TTY columns stay in lock step.
func WebhookFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Webhook] {
	p := format.New[backend.Webhook](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Webhook]{Name: "id", Header: "ID", Extract: func(h backend.Webhook) any { return h.ID }})
	p.AddField(format.Field[backend.Webhook]{Name: "url", Header: "URL", Extract: func(h backend.Webhook) any { return h.URL }})
	p.AddField(format.Field[backend.Webhook]{
		Name: "active", Header: "ACTIVE",
		Extract:   func(h backend.Webhook) any { return h.Active },
		ColorFunc: WebhookActiveColor(f.IOStreams),
	})
	// TTY column joins events with comma; JSON keeps them as an array (Field
	// returns the slice directly when the value is JSONOnly OR when serializing
	// to JSON — the format printer flattens slices to comma-joined strings on
	// the TTY path automatically.
	p.AddField(format.Field[backend.Webhook]{Name: "events", Header: "EVENTS", Extract: func(h backend.Webhook) any { return h.Events }})
	return p
}

// ParseEvents splits a comma-separated string of event keys into a deduped,
// trimmed slice. Empty entries (e.g. trailing commas) are skipped. Returns nil
// for an empty input so callers can rely on len(out)==0.
func ParseEvents(raw string) []string {
	if raw == "" {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, 4)
	for _, ev := range strings.Split(raw, ",") {
		ev = strings.TrimSpace(ev)
		if ev == "" {
			continue
		}
		if _, dup := seen[ev]; dup {
			continue
		}
		seen[ev] = struct{}{}
		out = append(out, ev)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
