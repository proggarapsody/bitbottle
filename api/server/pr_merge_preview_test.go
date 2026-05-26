package server_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerClient_DryRunMergePR_CanMerge(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"canMerge":true,"vetoes":[]}`)
	})
	result, err := client.DryRunMergePR("myproj", "my-repo", 42, "")
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-repo/pull-requests/42/merge/dry-run", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.True(t, result.CanMerge)
	assert.Empty(t, result.Vetoes)
}

func TestServerClient_DryRunMergePR_UppercasesProjectKey(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"canMerge":true,"vetoes":[]}`)
	})
	_, err := client.DryRunMergePR("lowercase", "my-repo", 1, "")
	require.NoError(t, err)
	assert.True(t, strings.Contains(gotPath, "/LOWERCASE/"))
}

func TestServerClient_DryRunMergePR_ReturnsVetoes(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"canMerge":false,"vetoes":[{"summaryMessage":"Not all required tasks are complete","detailedMessage":"Complete 2 tasks before merging"}]}`)
	})
	result, err := client.DryRunMergePR("PROJ", "repo", 5, "")
	require.NoError(t, err)
	assert.False(t, result.CanMerge)
	require.Len(t, result.Vetoes, 1)
	assert.Equal(t, "Not all required tasks are complete", result.Vetoes[0].SummaryMessage)
	assert.Equal(t, "Complete 2 tasks before merging", result.Vetoes[0].DetailMessage)
}

func TestServerClient_DryRunMergePR_404_ReturnsNotFound(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"errors":[{"message":"Pull request 99 does not exist"}]}`)
	})
	_, err := client.DryRunMergePR("PROJ", "repo", 99, "")
	require.Error(t, err)
}
