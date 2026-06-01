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

// ciAdapter normalizes Server and Cloud code-insights clients to a common
// interface used by the report and annotation commands.
type ciAdapter struct {
	server backend.CodeInsightsClient
	cloud  backend.CloudCodeInsightsClient
}

// resolveCIAdapter tries Server first, then Cloud. Returns (nil, err) only if
// both backends reject the host.
func resolveCIAdapter(c backend.Client, host string) (*ciAdapter, error) {
	if s, err := backend.AsCodeInsightsClient(c, host); err == nil {
		return &ciAdapter{server: s}, nil
	}
	if cl, err := backend.AsCloudCodeInsightsClient(c, host); err == nil {
		return &ciAdapter{cloud: cl}, nil
	}
	// Return Server's error — it carries the correct feature label for
	// non-Cloud hosts; Cloud error would be misleading for Server users.
	_, err := backend.AsCodeInsightsClient(c, host)
	return nil, err
}

func (a *ciAdapter) ListReports(project, slug, hash string) ([]backend.CodeInsightsReport, error) {
	if a.server != nil {
		return a.server.ListReports(project, slug, hash)
	}
	return a.cloud.ListCodeInsightsReports(project, slug, hash)
}

func (a *ciAdapter) GetReport(project, slug, hash, key string) (backend.CodeInsightsReport, error) {
	if a.server != nil {
		return a.server.GetReport(project, slug, hash, key)
	}
	return a.cloud.GetCodeInsightsReport(project, slug, hash, key)
}

func (a *ciAdapter) SetReport(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
	if a.server != nil {
		return a.server.SetReport(project, slug, hash, key, in)
	}
	return a.cloud.PutCodeInsightsReport(project, slug, hash, key, in)
}

func (a *ciAdapter) DeleteReport(project, slug, hash, key string) error {
	if a.server != nil {
		return a.server.DeleteReport(project, slug, hash, key)
	}
	return a.cloud.DeleteCodeInsightsReport(project, slug, hash, key)
}

func (a *ciAdapter) ListAnnotations(project, slug, hash, key string) ([]backend.CodeInsightsAnnotation, error) {
	if a.server != nil {
		return a.server.ListAnnotations(project, slug, hash, key)
	}
	return a.cloud.ListCodeInsightsAnnotations(project, slug, hash, key)
}

func (a *ciAdapter) AddAnnotations(project, slug, hash, key string, in []backend.CodeInsightsAnnotationInput) error {
	if a.server != nil {
		return a.server.AddAnnotations(project, slug, hash, key, in)
	}
	return a.cloud.PutCodeInsightsAnnotations(project, slug, hash, key, in)
}

func (a *ciAdapter) DeleteAnnotations(project, slug, hash, key string) error {
	if a.server != nil {
		return a.server.DeleteAnnotations(project, slug, hash, key)
	}
	// Cloud does not support annotation bulk-delete; return a typed error.
	return &backend.DomainError{
		Kind:    backend.ErrUnsupportedOnHost,
		Code:    backend.CodeHostUnsupported,
		Host:    "bitbucket.org",
		Feature: string(backend.FeatureCloudCodeInsights),
		Message: "annotation delete is not supported on Bitbucket Cloud",
	}
}
