package acceptancecloud

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_SanitizesCredentialsFromReturnedAndPersistedReport(t *testing.T) {
	bootstrapToken := "bootstrap-secret-sentinel"
	actorToken := "actor-secret-sentinel"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+bootstrapToken {
			t.Errorf("Authorization header = %q, want bootstrap credential", r.Header.Get("Authorization"))
		}
		// The redirect target becomes part of url.Error. Put both fake
		// credentials there to exercise sanitization of the returned error.
		w.Header().Set("Location", "/"+bootstrapToken+"/"+actorToken+"/redirect")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	dir := t.TempDir()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	baseURL := server.URL + "/" + bootstrapToken + "/" + actorToken
	report, runErr := Run(context.Background(), Config{
		BootstrapToken: bootstrapToken,
		ActorToken:     actorToken,
		Email:          "qa@example.test",
		Workspace:      "bitbottle-qa-lab",
		Project:        "BBQA",
		ReportDir:      dir,
		Binary:         binary,
		BaseURL:        baseURL,
	})
	if runErr == nil || report.Status != "infrastructure_failure" {
		t.Fatalf("Run() = status %q, error %v; want an infrastructure failure", report.Status, runErr)
	}

	persisted, err := os.ReadFile(filepath.Join(dir, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := []struct {
		name string
		text string
	}{
		{name: "returned error", text: runErr.Error()},
		{name: "returned report error", text: report.Error},
		{name: "persisted report", text: string(persisted)},
	}
	var redirectErr *url.Error
	if !errors.As(runErr, &redirectErr) || redirectErr.Err.Error() != "API redirect refused" {
		t.Errorf("Run() error = %v, want an unwrap-accessible redirect error", runErr)
	}
	if !strings.Contains(runErr.Error(), "API redirect refused") {
		t.Errorf("returned error lost useful transport context: %q", runErr)
	}
	for _, artifact := range artifacts {
		if !strings.Contains(artifact.text, "API redirect refused") {
			t.Errorf("%s lost useful transport error context: %q", artifact.name, artifact.text)
		}
		for _, secret := range []string{bootstrapToken, actorToken} {
			if strings.Contains(artifact.text, secret) {
				t.Errorf("%s leaked credential %q: %q", artifact.name, secret, artifact.text)
			}
		}
	}
}
