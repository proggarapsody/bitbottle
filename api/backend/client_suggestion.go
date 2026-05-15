package backend

import "fmt"

// SuggestionApplyResult is returned when a suggested change is successfully applied.
type SuggestionApplyResult struct {
	CommitHash    string
	CommitMessage string
}

// SuggestionApplier applies Bitbucket Data Center inline code suggestions.
// Bitbucket Cloud has no equivalent primitive, so callers route through
// AsSuggestionApplier to surface the constraint as a typed ErrUnsupportedOnHost.
type SuggestionApplier interface {
	// ApplySuggestion commits a suggested change to the PR source branch.
	// The adapter is responsible for fetching the current PR version for
	// optimistic locking before issuing the POST.
	ApplySuggestion(ns, slug string, prID, commentID, suggestionID int) (SuggestionApplyResult, error)

	// GetSuggestionPreview returns the suggestion body text from the comment
	// without applying the change.
	GetSuggestionPreview(ns, slug string, prID, commentID int) (string, error)
}

// FeaturePRSuggestion names the PR suggestion capability.
const FeaturePRSuggestion Feature = "pr-suggestion"

// AsSuggestionApplier returns the SuggestionApplier view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when the backend has no suggestion
// apply primitive (currently Bitbucket Cloud).
func AsSuggestionApplier(c Client, host string) (SuggestionApplier, error) {
	r, ok := c.(SuggestionApplier)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePRSuggestion),
			Message: fmt.Sprintf("pr suggestion apply is not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return r, nil
}
