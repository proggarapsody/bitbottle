// Package shared holds format helpers for variable subcommands.
package shared

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// SecuredPlaceholder is rendered in TTY tables in place of a secured variable's value.
const SecuredPlaceholder = "<secured>"

// DisplayVariableValue redacts secured pipeline variable values for TTY output.
func DisplayVariableValue(v backend.PipelineVariable) string {
	if v.Secured {
		return SecuredPlaceholder
	}
	return v.Value
}

// VariableFields constructs the formatter for repository and workspace scope
// variable lists (PipelineVariable). Secured values are redacted via
// DisplayVariableValue in both TTY and JSON output paths.
func VariableFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.PipelineVariable] {
	p := format.New[backend.PipelineVariable](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.PipelineVariable]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(v backend.PipelineVariable) any { return v.UUID }})
	p.AddField(format.Field[backend.PipelineVariable]{Name: "key", Header: "KEY", Extract: func(v backend.PipelineVariable) any { return v.Key }})
	p.AddField(format.Field[backend.PipelineVariable]{
		Name: "value", Header: "VALUE",
		Extract: func(v backend.PipelineVariable) any { return DisplayVariableValue(v) },
	})
	p.AddField(format.Field[backend.PipelineVariable]{Name: "secured", Header: "SECURED", Extract: func(v backend.PipelineVariable) any { return v.Secured }})
	return p
}

// DisplayEnvVariableValue redacts secured env variable values for TTY output.
func DisplayEnvVariableValue(v backend.EnvVariable) string {
	if v.Secured {
		return SecuredPlaceholder
	}
	return v.Value
}

// EnvVariableFields constructs the formatter for deployment scope variable
// lists (EnvVariable). Secured values are redacted in both TTY and JSON paths.
func EnvVariableFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.EnvVariable] {
	p := format.New[backend.EnvVariable](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.EnvVariable]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(v backend.EnvVariable) any { return v.UUID }})
	p.AddField(format.Field[backend.EnvVariable]{Name: "key", Header: "KEY", Extract: func(v backend.EnvVariable) any { return v.Key }})
	p.AddField(format.Field[backend.EnvVariable]{
		Name: "value", Header: "VALUE",
		Extract: func(v backend.EnvVariable) any { return DisplayEnvVariableValue(v) },
	})
	p.AddField(format.Field[backend.EnvVariable]{Name: "secured", Header: "SECURED", Extract: func(v backend.EnvVariable) any { return v.Secured }})
	return p
}
