package server_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

// jsonErrServer returns a Bitbucket Server-shaped JSON error envelope at
// the given status. The Server adapter's decoder reads the first error's
// `message` field; matching that shape here keeps the test minimal while
// exercising the same parse path as production.
func jsonErrServer(t *testing.T, status int, message string) *httptest.Server {
	t.Helper()
	body := `{"errors":[{"message":"` + message + `"}]}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestServerClient_GetRepo_404_StampsRepoNotFoundCode(t *testing.T) {
	srv := jsonErrServer(t, 404, "Repository does not exist.")
	c := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := c.GetRepo("MYPROJ", "missing")
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
	if de.ID != "MYPROJ/missing" {
		t.Errorf("ID = %q, want %q", de.ID, "MYPROJ/missing")
	}
}

func TestServerClient_GetPR_404_StampsPRNotFoundCode(t *testing.T) {
	srv := jsonErrServer(t, 404, "PR does not exist.")
	c := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := c.GetPR("MYPROJ", "repo", 42)
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

func TestServerClient_DeleteBranch_403_StampsBranchProtectedCode(t *testing.T) {
	srv := jsonErrServer(t, 403, "Branch is protected.")
	c := server.NewClient(srv.Client(), srv.URL, "tok", "")
	err := c.DeleteBranch("MYPROJ", "repo", "main")
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
