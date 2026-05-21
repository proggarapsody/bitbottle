package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func strPtr(s string) *string { return &s }

func TestServerClient_EditRepo_DescriptionUpdate(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"slug":"my-service","name":"my-service","scmId":"git","project":{"key":"MYPROJ"},"links":{"self":[{"href":"https://bb/projects/MYPROJ/repos/my-service/browse"}]}}`))
	})

	desc := "new description"
	repo, err := client.EditRepo("MYPROJ", "my-service", backend.EditRepoInput{
		Description: strPtr(desc),
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service", gotPath)
	assert.Equal(t, desc, gotBody["description"])
	assert.Equal(t, "my-service", repo.Slug)
	assert.Equal(t, "MYPROJ", repo.Namespace)
}

func TestServerClient_EditRepo_CloudOnlyFieldsIgnored(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"slug":"my-service","name":"my-service","scmId":"git","project":{"key":"MYPROJ"},"links":{"self":[{"href":"https://bb/projects/MYPROJ/repos/my-service/browse"}]}}`))
	})

	website := "https://example.com"
	language := "Go"
	_, err := client.EditRepo("MYPROJ", "my-service", backend.EditRepoInput{
		Website:  strPtr(website),
		Language: strPtr(language),
	})
	require.NoError(t, err)
	// Cloud-only fields must not be sent to Server
	_, hasWebsite := gotBody["website"]
	assert.False(t, hasWebsite, "website should not be sent to Server")
	_, hasLanguage := gotBody["language"]
	assert.False(t, hasLanguage, "language should not be sent to Server")
	// description is nil — must not appear
	_, hasDesc := gotBody["description"]
	assert.False(t, hasDesc, "nil description should not appear in body")
}
