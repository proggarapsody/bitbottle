package backend_test

import (
	"errors"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// emptyClient is a Client implementation that satisfies only the embedded
// composite Client interface — no optional capabilities — so every As*
// helper falls into the type-assertion failure branch.
type emptyClient struct{ backend.Client }

// TestAsPipelineClient_StampsHostUnsupportedCode verifies that the As*
// helpers stamp CodeHostUnsupported, not just the bare ErrUnsupportedOnHost
// kind. The renderer's host.unsupported template needs the Code to pick the
// catalogue entry; without it, the kind-fallback line ("X is not available
// on Y") fires and the catalogue's hint is lost.
func TestAsPipelineClient_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsPipelineClient(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeaturePipelines))
}

func TestAsRepoForker_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsRepoForker(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeatureRepoFork))
}

func TestAsWorkspaceClient_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsWorkspaceClient(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeatureWorkspaces))
}

func TestAsIssueClient_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsIssueClient(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeatureIssues))
}

func TestAsBranchProtector_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsBranchProtector(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeatureBranchProtect))
}

func TestAsCodeSearcher_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsCodeSearcher(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeatureCodeSearch))
}

func TestAsDefaultReviewersResolver_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsDefaultReviewersResolver(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeatureDefaultReviewers))
}

func TestAsPRCommentResolver_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsPRCommentResolver(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeaturePRCommentResolve))
}

func TestAsCodeInsightsClient_StampsHostUnsupportedCode(t *testing.T) {
	_, err := backend.AsCodeInsightsClient(emptyClient{}, "h.example")
	requireUnsupportedHostCode(t, err, string(backend.FeatureCodeInsights))
}

func requireUnsupportedHostCode(t *testing.T, err error, wantFeature string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != backend.CodeHostUnsupported {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodeHostUnsupported)
	}
	if de.Feature != wantFeature {
		t.Errorf("Feature = %q, want %q", de.Feature, wantFeature)
	}
	if de.Host != "h.example" {
		t.Errorf("Host = %q, want %q", de.Host, "h.example")
	}
}
