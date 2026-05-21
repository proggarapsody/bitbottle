package backend

import (
	"fmt"
	"time"
)

// SnippetClient is implemented only by Bitbucket Cloud clients. Bitbucket
// Server / Data Center has no snippets API, so the entire snippet surface is
// gated behind AsSnippetClient.
type SnippetClient interface {
	ListSnippets(workspace string, limit int) ([]Snippet, error)
	GetSnippet(workspace, id string) (Snippet, error)
	CreateSnippet(workspace string, in CreateSnippetInput) (Snippet, error)
	DeleteSnippet(workspace, id string) error
}

// FeatureSnippets names the snippets capability for typed-error reporting.
const FeatureSnippets Feature = "snippets"

// AsSnippetClient returns the SnippetClient view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsSnippetClient(c Client, host string) (SnippetClient, error) {
	sc, ok := c.(SnippetClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureSnippets),
			Message: fmt.Sprintf("snippets are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return sc, nil
}

// Snippet is the domain representation of a Bitbucket Cloud snippet.
type Snippet struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Owner     string        `json:"owner"`
	IsPrivate bool          `json:"is_private"`
	CreatedOn time.Time     `json:"created_on"`
	Files     []SnippetFile `json:"files"`
	WebURL    string        `json:"web_url"`
}

// SnippetFile is a file attached to a snippet.
type SnippetFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// CreateSnippetInput holds the parameters for creating a new snippet.
type CreateSnippetInput struct {
	Title     string
	IsPrivate bool
	Files     []SnippetFile
}
