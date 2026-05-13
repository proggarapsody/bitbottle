// Package shared holds helpers used across pipeline subcommands.
package shared

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// Note: variable formatting helpers (VariableFields, DisplayVariableValue,
// SecuredPlaceholder) were removed — use pkg/cmd/variable/shared instead.

// PipelineStateColor maps Bitbucket Cloud pipeline / step state strings to
// colors. Both Pipeline and PipelineStep share the same state vocabulary, so
// they share this helper. SUCCESSFUL is green; FAILED / ERROR / STOPPED are
// red; IN_PROGRESS / PENDING are yellow. Unknown states pass through.
//
// Exported so the few cmd packages outside `shared` (none today, but
// `pipeline view` could grow that need) can reuse the mapping without
// re-deriving it.
func PipelineStateColor(ios *iostreams.IOStreams) func(string) string {
	return func(state string) string {
		switch state {
		case "SUCCESSFUL":
			return ios.ColorGreen(state)
		case "FAILED", "ERROR", "STOPPED":
			return ios.ColorRed(state)
		case "IN_PROGRESS", "PENDING":
			return ios.ColorYellow(state)
		default:
			return state
		}
	}
}

// PipelineFields constructs the formatter used by both `pipeline list` and
// `pipeline view` so the JSON field names and TTY column layout stay in lock
// step.
func PipelineFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Pipeline] {
	p := format.New[backend.Pipeline](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Pipeline]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(pl backend.Pipeline) any { return pl.UUID }})
	p.AddField(format.Field[backend.Pipeline]{Name: "buildNumber", Header: "BUILD", Extract: func(pl backend.Pipeline) any { return pl.BuildNumber }})
	p.AddField(format.Field[backend.Pipeline]{
		Name: "state", Header: "STATE",
		Extract:   func(pl backend.Pipeline) any { return pl.State },
		ColorFunc: PipelineStateColor(f.IOStreams),
	})
	p.AddField(format.Field[backend.Pipeline]{Name: "refName", Header: "BRANCH/TAG", Extract: func(pl backend.Pipeline) any { return pl.RefName }})
	p.AddField(format.Field[backend.Pipeline]{Name: "duration", Header: "DURATION", Extract: func(pl backend.Pipeline) any { return pl.Duration }})
	p.AddField(format.Field[backend.Pipeline]{Name: "webURL", Header: "URL", JSONOnly: true, Extract: func(pl backend.Pipeline) any { return pl.WebURL }})
	return p
}

// StepFields constructs the formatter for `pipeline steps`.
func StepFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.PipelineStep] {
	p := format.New[backend.PipelineStep](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.PipelineStep]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(s backend.PipelineStep) any { return s.UUID }})
	p.AddField(format.Field[backend.PipelineStep]{Name: "name", Header: "NAME", Extract: func(s backend.PipelineStep) any { return s.Name }})
	p.AddField(format.Field[backend.PipelineStep]{
		Name: "state", Header: "STATE",
		Extract:   func(s backend.PipelineStep) any { return s.State },
		ColorFunc: PipelineStateColor(f.IOStreams),
	})
	p.AddField(format.Field[backend.PipelineStep]{Name: "result", Header: "RESULT", JSONOnly: true, Extract: func(s backend.PipelineStep) any { return s.Result }})
	p.AddField(format.Field[backend.PipelineStep]{Name: "duration", Header: "DURATION", Extract: func(s backend.PipelineStep) any { return s.Duration }})
	return p
}

