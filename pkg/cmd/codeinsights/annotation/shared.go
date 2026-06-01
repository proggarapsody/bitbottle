package annotation

import (
	"github.com/proggarapsody/bitbottle/api/backend"
)

// ciAdapter normalizes Server and Cloud code-insights clients to a common
// interface used by the annotation commands.
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
