package cloud_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// jsonErrServer returns a Cloud-shaped JSON error envelope at status,
// so the adapter's decodeErrorMessage produces a non-empty Message that
// adapters can pattern-match on.
func jsonErrServer(t *testing.T, status int, message string) *httptest.Server {
	t.Helper()
	body := `{"type":"error","error":{"message":"` + message + `"}}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestCloudClient_GetRepo_404_StampsRepoNotFoundCode verifies that a 404 on
// GetRepo carries CodeRepoNotFound, Resource="repository", and ID="ns/slug".
// Without the stamp, the renderer falls back to the generic ErrNotFound line,
// which doesn't tell users to check casing or use the project key on Server.
func TestCloudClient_GetRepo_404_StampsRepoNotFoundCode(t *testing.T) {
	srv := jsonErrServer(t, 404, "Repository not found.")
	c := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := c.GetRepo("ws", "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != backend.CodeRepoNotFound {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodeRepoNotFound)
	}
	if de.Resource != "repository" {
		t.Errorf("Resource = %q, want %q", de.Resource, "repository")
	}
	if de.ID != "ws/missing" {
		t.Errorf("ID = %q, want %q", de.ID, "ws/missing")
	}
}

// TestCloudClient_GetPR_404_StampsPRNotFoundCode same idea for PRs.
func TestCloudClient_GetPR_404_StampsPRNotFoundCode(t *testing.T) {
	srv := jsonErrServer(t, 404, "No such pull request.")
	c := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := c.GetPR("ws", "repo", 42)
	if err == nil {
		t.Fatal("expected error")
	}
	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != backend.CodePRNotFound {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodePRNotFound)
	}
	if de.ID != "42" {
		t.Errorf("ID = %q, want %q", de.ID, "42")
	}
}

// TestCloudClient_MergePR_409_Conflict_StampsConflictCode verifies that a
// 409 with a "conflict"-shaped message stamps CodePRMergeConflict.
func TestCloudClient_MergePR_409_Conflict_StampsConflictCode(t *testing.T) {
	srv := jsonErrServer(t, 409, "There are merge conflicts that must be resolved before this pull request can be merged.")
	c := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := c.MergePR("ws", "repo", 42, backend.MergePRInput{})
	if err == nil {
		t.Fatal("expected error")
	}
	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != backend.CodePRMergeConflict {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodePRMergeConflict)
	}
}

// TestCloudClient_CreatePR_409_StampsDuplicateBranchCode verifies that a
// 409 on CreatePR is stamped as a duplicate-branch case.
func TestCloudClient_CreatePR_409_StampsDuplicateBranchCode(t *testing.T) {
	srv := jsonErrServer(t, 409, "There is already an open pull request for these branches.")
	c := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := c.CreatePR("ws", "repo", backend.CreatePRInput{
		Title:      "x",
		FromBranch: "feat/x",
		ToBranch:   "main",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != backend.CodePRCreateDuplicateBranch {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodePRCreateDuplicateBranch)
	}
}

// TestCloudClient_CreatePR_400_ReviewerUnknown_StampsReviewerCode verifies
// that 400 with reviewer-shaped substring stamps CodePRReviewerUnknown.
func TestCloudClient_CreatePR_400_ReviewerUnknown_StampsReviewerCode(t *testing.T) {
	srv := jsonErrServer(t, 400, "reviewers: User with username 'nonesuch' does not exist.")
	c := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := c.CreatePR("ws", "repo", backend.CreatePRInput{
		Title:      "x",
		FromBranch: "feat/x",
		ToBranch:   "main",
		Reviewers:  []string{"nonesuch"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != backend.CodePRReviewerUnknown {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodePRReviewerUnknown)
	}
}

// TestCloudClient_DeleteBranch_403_StampsBranchProtectedCode verifies that
// 403 on DeleteBranch is treated as branch protection rather than generic
// permission failure.
func TestCloudClient_DeleteBranch_403_StampsBranchProtectedCode(t *testing.T) {
	srv := jsonErrServer(t, 403, "Branch is protected and cannot be deleted.")
	c := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := c.DeleteBranch("ws", "repo", "main")
	if err == nil {
		t.Fatal("expected error")
	}
	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if de.Code != backend.CodeBranchProtected {
		t.Errorf("Code = %q, want %q", de.Code, backend.CodeBranchProtected)
	}
}
