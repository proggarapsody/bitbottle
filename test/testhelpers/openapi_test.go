package testhelpers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// goodServerPR returns a map that satisfies all required fields of RestPullRequest.
func goodServerPR() map[string]any {
	return map[string]any{
		"id":          1,
		"version":     0,
		"title":       "t",
		"description": "d",
		"state":       "OPEN",
		"draft":       false,
		"author": map[string]any{
			"user": map[string]any{
				"name":        "alice",
				"slug":        "alice",
				"displayName": "Alice",
			},
		},
		"fromRef": map[string]any{
			"id":           "refs/heads/feature",
			"displayId":    "feature",
			"latestCommit": "abc123",
		},
		"toRef": map[string]any{
			"id":           "refs/heads/main",
			"displayId":    "main",
			"latestCommit": "def456",
		},
		"reviewers": []any{},
		"links": map[string]any{
			"self": []any{
				map[string]any{"href": "http://x"},
			},
		},
	}
}

func TestOpenAPI_ServerRestPullRequest_GoodPassesValidation(t *testing.T) {
	ValidateAgainstSchema(t, "RestPullRequest", goodServerPR(), "server")
}

func TestOpenAPI_ServerRestPullRequest_MissingTitleFails(t *testing.T) {
	bad := goodServerPR()
	delete(bad, "title")

	err := ValidateAgainstSchemaErr("RestPullRequest", bad, "server")
	require.NotNil(t, err, "expected validation error for missing required field 'title'")
}

func TestOpenAPI_CloudPullRequest_GoodPassesValidation(t *testing.T) {
	good := map[string]any{
		"id":          42,
		"title":       "cloud pr",
		"description": "desc",
		"state":       "OPEN",
		"draft":       false,
		"author": map[string]any{
			"account_id":   "acc-123",
			"nickname":     "bob",
			"display_name": "Bob",
		},
		// CloudBranchRef requires both branch and commit.
		"source": map[string]any{
			"branch": map[string]any{"name": "feature"},
			"commit": map[string]any{"hash": "abc123"},
		},
		// CloudDestinationRef requires only branch.
		"destination": map[string]any{
			"branch": map[string]any{"name": "main"},
		},
		"links": map[string]any{
			"html": map[string]any{"href": "http://bitbucket.org/x"},
		},
		"reviewers": []any{},
	}
	ValidateAgainstSchema(t, "CloudPullRequest", good, "cloud")
}

func TestOpenAPI_CloudPullRequest_MissingTitleFails(t *testing.T) {
	bad := map[string]any{
		"id":          42,
		"description": "desc",
		"state":       "OPEN",
		"draft":       false,
		"author": map[string]any{
			"account_id": "acc-123",
		},
		"source":      map[string]any{"branch": map[string]any{"name": "feature"}},
		"destination": map[string]any{"branch": map[string]any{"name": "main"}},
		"links":       map[string]any{"html": map[string]any{"href": "http://x"}},
		"reviewers":   []any{},
	}

	err := ValidateAgainstSchemaErr("CloudPullRequest", bad, "cloud")
	require.NotNil(t, err, "expected validation error for missing required field 'title'")
}

func TestOpenAPI_UnknownBackendFails(t *testing.T) {
	err := ValidateAgainstSchemaErr("RestPullRequest", goodServerPR(), "unknown-backend")
	require.NotNil(t, err, "expected error for unknown backend")
}

func TestOpenAPI_UnknownSchemaFails(t *testing.T) {
	err := ValidateAgainstSchemaErr("NonExistentSchema", goodServerPR(), "server")
	require.NotNil(t, err, "expected error for unknown schema name")
}
