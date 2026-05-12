package report

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// reportFields builds the Printer for Code Insights reports.
func reportFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.CodeInsightsReport] {
	p := format.New[backend.CodeInsightsReport](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.CodeInsightsReport]{
		Name: "key", Header: "KEY",
		Extract: func(r backend.CodeInsightsReport) any { return r.Key },
	})
	p.AddField(format.Field[backend.CodeInsightsReport]{
		Name: "title", Header: "TITLE",
		Extract: func(r backend.CodeInsightsReport) any { return r.Title },
	})
	p.AddField(format.Field[backend.CodeInsightsReport]{
		Name: "result", Header: "RESULT",
		Extract: func(r backend.CodeInsightsReport) any { return r.Result },
	})
	p.AddField(format.Field[backend.CodeInsightsReport]{
		Name: "report_type", Header: "TYPE",
		Extract: func(r backend.CodeInsightsReport) any { return r.ReportType },
	})
	p.AddField(format.Field[backend.CodeInsightsReport]{
		Name: "reporter", Header: "REPORTER",
		Extract: func(r backend.CodeInsightsReport) any { return r.Reporter },
	})
	p.AddField(format.Field[backend.CodeInsightsReport]{
		Name: "details", Header: "DETAILS",
		JSONOnly: true,
		Extract:  func(r backend.CodeInsightsReport) any { return r.Details },
	})
	p.AddField(format.Field[backend.CodeInsightsReport]{
		Name: "link", Header: "LINK",
		JSONOnly: true,
		Extract:  func(r backend.CodeInsightsReport) any { return r.Link },
	})
	return p
}
