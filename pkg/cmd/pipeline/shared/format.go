// Package shared holds helpers used across pipeline subcommands.
package shared

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// PipelineFields constructs the formatter used by both `pipeline list` and
// `pipeline view` so the JSON field names and TTY column layout stay in lock
// step.
func PipelineFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Pipeline] {
	p := format.New[backend.Pipeline](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Pipeline]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(pl backend.Pipeline) any { return pl.UUID }})
	p.AddField(format.Field[backend.Pipeline]{Name: "buildNumber", Header: "BUILD", Extract: func(pl backend.Pipeline) any { return pl.BuildNumber }})
	p.AddField(format.Field[backend.Pipeline]{Name: "state", Header: "STATE", Extract: func(pl backend.Pipeline) any { return pl.State }})
	p.AddField(format.Field[backend.Pipeline]{Name: "refName", Header: "BRANCH/TAG", Extract: func(pl backend.Pipeline) any { return pl.RefName }})
	p.AddField(format.Field[backend.Pipeline]{Name: "duration", Header: "DURATION", Extract: func(pl backend.Pipeline) any { return pl.Duration }})
	p.AddField(format.Field[backend.Pipeline]{Name: "webURL", Header: "URL", JSONOnly: true, Extract: func(pl backend.Pipeline) any { return pl.WebURL }})
	return p
}
