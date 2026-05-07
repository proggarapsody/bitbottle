// Package shared holds helpers used across pipeline subcommands.
package shared

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

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
func PipelineFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Pipeline] {
	p := format.New[backend.Pipeline](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
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
func StepFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.PipelineStep] {
	p := format.New[backend.PipelineStep](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
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

// SecuredPlaceholder is rendered in TTY tables in place of a secured pipeline
// variable's value. The JSON path emits the raw (empty) value so scripts get a
// boolean `secured` field instead of a fake placeholder.
const SecuredPlaceholder = "<secured>"

// DisplayVariableValue is the TTY-only redaction helper. Pure: takes a domain
// value, returns the string the column should print.
func DisplayVariableValue(v backend.PipelineVariable) string {
	if v.Secured {
		return SecuredPlaceholder
	}
	return v.Value
}

// VariableFields constructs the formatter for `pipeline variable list`. The
// TTY value column applies DisplayVariableValue so secured variables print the
// placeholder; the JSON `value` field passes through the raw value (always
// empty for secured variables, since the API never returns those).
func VariableFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.PipelineVariable] {
	p := format.New[backend.PipelineVariable](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.PipelineVariable]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(v backend.PipelineVariable) any { return v.UUID }})
	p.AddField(format.Field[backend.PipelineVariable]{Name: "key", Header: "KEY", Extract: func(v backend.PipelineVariable) any { return v.Key }})
	// Both TTY and JSON paths route through DisplayVariableValue so secured
	// values cannot be exfiltrated by switching output modes.
	p.AddField(format.Field[backend.PipelineVariable]{
		Name: "value", Header: "VALUE",
		Extract: func(v backend.PipelineVariable) any { return DisplayVariableValue(v) },
	})
	p.AddField(format.Field[backend.PipelineVariable]{Name: "secured", Header: "SECURED", Extract: func(v backend.PipelineVariable) any { return v.Secured }})
	return p
}
