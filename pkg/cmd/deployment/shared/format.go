// Package shared holds format helpers for deployment and environment subcommands.
package shared

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// DeploymentStateColor maps Bitbucket Cloud deployment state strings to colors.
func DeploymentStateColor(ios *iostreams.IOStreams) func(string) string {
	return func(state string) string {
		switch state {
		case "COMPLETED":
			return ios.ColorGreen(state)
		case "FAILED", "STOPPED":
			return ios.ColorRed(state)
		case "IN_PROGRESS", "PENDING":
			return ios.ColorYellow(state)
		default:
			return state
		}
	}
}

// DeploymentFields constructs the formatter for deployment list/view.
func DeploymentFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Deployment] {
	p := format.New[backend.Deployment](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Deployment]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(d backend.Deployment) any { return d.UUID }})
	p.AddField(format.Field[backend.Deployment]{
		Name: "state", Header: "STATE",
		Extract:   func(d backend.Deployment) any { return d.State },
		ColorFunc: DeploymentStateColor(f.IOStreams),
	})
	p.AddField(format.Field[backend.Deployment]{Name: "environment", Header: "ENVIRONMENT", Extract: func(d backend.Deployment) any { return d.Environment.Name }})
	p.AddField(format.Field[backend.Deployment]{Name: "release", Header: "RELEASE", Extract: func(d backend.Deployment) any { return d.Release.Name }})
	p.AddField(format.Field[backend.Deployment]{Name: "commitHash", Header: "COMMIT", Extract: func(d backend.Deployment) any {
		h := d.Release.CommitHash
		if len(h) > 7 {
			h = h[:7]
		}
		return h
	}})
	return p
}

// EnvironmentFields constructs the formatter for environment list.
func EnvironmentFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.Environment] {
	p := format.New[backend.Environment](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.Environment]{Name: "uuid", Header: "UUID", JSONOnly: true, Extract: func(e backend.Environment) any { return e.UUID }})
	p.AddField(format.Field[backend.Environment]{Name: "name", Header: "NAME", Extract: func(e backend.Environment) any { return e.Name }})
	p.AddField(format.Field[backend.Environment]{Name: "type", Header: "TYPE", Extract: func(e backend.Environment) any { return e.Type }})
	p.AddField(format.Field[backend.Environment]{Name: "rank", Header: "RANK", Extract: func(e backend.Environment) any { return e.Rank }})
	return p
}

// SecuredPlaceholder is rendered in place of a secured env variable value.
const SecuredPlaceholder = "<secured>"

// DisplayEnvVariableValue redacts secured env variable values for TTY output.
func DisplayEnvVariableValue(v backend.EnvVariable) string {
	if v.Secured {
		return SecuredPlaceholder
	}
	return v.Value
}

// EnvVariableFields constructs the formatter for environment variable list.
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
