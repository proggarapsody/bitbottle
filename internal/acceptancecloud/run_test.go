package acceptancecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcptransport "github.com/mark3labs/mcp-go/client/transport"
)

func TestRun_RejectsMissingCredentialAndWritesSanitizedReport(t *testing.T) {
	dir := t.TempDir()
	report, err := Run(context.Background(), Config{ReportDir: dir})
	if err == nil || report.Status != "infrastructure_failure" {
		t.Fatalf("Run() = status %q, error %v", report.Status, err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved Report
	if err = json.Unmarshal(b, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.FailCount != 1 || saved.RunCount != 1 || saved.Backend != "cloud" {
		t.Fatalf("unexpected report summary: %+v", saved)
	}
}

func TestCleanup_RefusesRepositoryWithoutOwnershipMarker(t *testing.T) {
	deleted := false
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleted = true
			http.Error(w, "unexpected", 500)
			return
		}
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer bootstrap-secret" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{owner}","description":"someone else's repository"}`))
	}))
	defer s.Close()
	m := Manifest{Version: 1, RunID: "0000000000001234", Resource: Resource{Workspace: "bitbottle-qa-lab", Slug: "bbqa-run-0000000000001234", UUID: "{owner}", Marker: "bitbottle-qa-run:0000000000001234"}}
	// Public Cleanup reads durable manifests; exercise the ownership boundary with a temp manifest.
	p := filepath.Join(t.TempDir(), "manifest.json")
	if e := m.Write(p); e != nil {
		t.Fatal(e)
	}
	err := Cleanup(context.Background(), p, "bootstrap-secret", s.URL)
	if err == nil || !strings.Contains(err.Error(), "marker mismatch") {
		t.Fatalf("Cleanup() error = %v, want ownership refusal", err)
	}
	if deleted {
		t.Fatal("Cleanup() deleted a repository with a different marker")
	}
}

func TestRun_SeedFailureCleansManifestOwnedRepository(t *testing.T) {
	for _, mode := range []struct {
		name                         string
		cleanupFails, obstructReport bool
	}{
		{name: "cleanup succeeds"},
		{name: "cleanup failure is red", cleanupFails: true},
		{name: "report persistence failure is red", obstructReport: true},
	} {
		t.Run(mode.name, func(t *testing.T) {
			created, declined, deleted, sawCreate := false, false, false, false
			marker := ""
			uuid := "{qa-repo}"
			dir := t.TempDir()
			reportPath := filepath.Join(dir, "report.json")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/repositories/bitbucket-qa-lab" || r.URL.Path == "/repositories/bitbucket-qa-lab/" {
					http.NotFound(w, r)
					return
				}
				if r.Header.Get("Authorization") != "Bearer bootstrap" {
					http.Error(w, "wrong credential", http.StatusUnauthorized)
					return
				}
				switch {
				case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repositories/") && strings.HasSuffix(r.URL.Path, "/commit/main"):
					_, _ = io.WriteString(w, `{"hash":"seed-hash"}`)
				case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repositories/") && r.URL.Query().Get("state") == "OPEN":
					_, _ = io.WriteString(w, `{"values":[{"id":55,"state":"OPEN"}]}`)
				case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repositories/"):
					if !created || deleted {
						http.NotFound(w, r)
						return
					}
					_, _ = fmt.Fprintf(w, `{"uuid":%q,"description":%q}`, uuid, marker)
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/pullrequests/55/decline"):
					declined = true
					w.WriteHeader(http.StatusOK)
				case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/repositories/") && !strings.Contains(r.URL.Path, "/src") && !strings.Contains(r.URL.Path, "/refs/"):
					sawCreate = true
					if r.Method != http.MethodPost {
						t.Errorf("repository create method = %s", r.Method)
					}
					var in struct {
						Description string `json:"description"`
					}
					if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
						t.Errorf("decode create: %v", err)
					}
					marker = in.Description
					created = true
					if mode.obstructReport {
						if err := os.Mkdir(reportPath, 0700); err != nil {
							t.Errorf("obstruct report path: %v", err)
						}
					}
					w.WriteHeader(http.StatusCreated)
					_, _ = fmt.Fprintf(w, `{"uuid":%q,"description":%q}`, uuid, marker)
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/src"):
					w.WriteHeader(http.StatusCreated)
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/refs/branches"):
					http.Error(w, "injected seed branch failure", http.StatusInternalServerError)
				case r.Method == http.MethodDelete:
					if mode.cleanupFails {
						http.Error(w, "injected delete failure", http.StatusInternalServerError)
						return
					}
					deleted = true
					w.WriteHeader(http.StatusNoContent)
				default:
					http.Error(w, fmt.Sprintf("unexpected %s %s", r.Method, r.URL.String()), http.StatusNotFound)
				}
			}))
			defer server.Close()
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			report, runErr := Run(context.Background(), Config{BootstrapToken: "bootstrap", ActorToken: "actor", Email: "qa@example.test", Workspace: "bitbottle-qa-lab", Project: "BBQA", ReportDir: dir, Binary: binary, BaseURL: server.URL})
			if runErr == nil {
				t.Fatal("Run() unexpectedly succeeded after injected seed failure")
			}
			if !sawCreate {
				t.Fatalf("repository creation did not use the expected POST endpoint: err=%v report=%+v", runErr, report)
			}
			if mode.cleanupFails {
				if report.Status != "cleanup_failure" || report.Cleanup.Status != "fail" || deleted {
					t.Fatalf("cleanup failure report=%+v deleted=%v", report, deleted)
				}
			} else if mode.obstructReport {
				if report.Status != "infrastructure_failure" || report.Checks["report_persistence"].Status != "fail" || report.Cleanup.Status != "pass" || !deleted {
					t.Fatalf("report persistence result report=%+v deleted=%v", report, deleted)
				}
				if !strings.Contains(runErr.Error(), "HTTP 500") || !strings.Contains(runErr.Error(), "report persistence failed") {
					t.Fatalf("returned error did not preserve product and evidence failures: %v", runErr)
				}
			} else if report.Status != "infrastructure_failure" || report.Cleanup.Status != "pass" || !declined || !deleted {
				t.Fatalf("cleanup result report=%+v declined=%v deleted=%v", report, declined, deleted)
			}
			if _, err = os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
				t.Fatalf("persisted manifest missing after failure: %v", err)
			}
		})
	}
}

func TestCleanup_RefusesRepositoryWithMismatchedUUID(t *testing.T) {
	deleted := false
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleted = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = io.WriteString(w, `{"uuid":"{other}","description":"bitbottle-qa-run:0000000000001234"}`)
	}))
	defer s.Close()
	m := Manifest{Version: 1, RunID: "0000000000001234", Resource: Resource{Workspace: "bitbottle-qa-lab", Slug: "bbqa-run-0000000000001234", UUID: "{expected}", Marker: "bitbottle-qa-run:0000000000001234"}}
	p := filepath.Join(t.TempDir(), "manifest.json")
	if e := m.Write(p); e != nil {
		t.Fatal(e)
	}
	err := Cleanup(context.Background(), p, "bootstrap-secret", s.URL)
	if err == nil || !strings.Contains(err.Error(), "UUID mismatch") {
		t.Fatalf("Cleanup() error = %v, want UUID refusal", err)
	}
	if deleted {
		t.Fatal("Cleanup() deleted a repository with a different UUID")
	}
}

func TestAPIClient_RefusesRedirects(t *testing.T) {
	called := false
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, http.StatusFound) }))
	defer source.Close()
	c, err := newAPI(source.URL, "secret-token")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.request(context.Background(), http.MethodGet, "/resource", nil, "", nil)
	if err == nil || !strings.Contains(err.Error(), "redirect refused") {
		t.Fatalf("request error = %v, want redirect refusal", err)
	}
	if called {
		t.Fatal("redirect target received a request")
	}
}

func TestFilteredEnv_DropsInheritedCloudAndPrivilegedValues(t *testing.T) {
	got := strings.Join(filteredEnv([]string{"PATH=/usr/bin", "BB_HOST=bitbucket.org", "BB_TOKEN=actor", "BB_CLOUD_BOOTSTRAP_TOKEN=bootstrap", "BB_CLOUD_BASE_URL=http://evil.test", "HTTPS_PROXY=http://proxy.test", "AWS_SECRET_ACCESS_KEY=secret", "HOME=/tmp/isolated"}), "\n")
	for _, forbidden := range []string{"bootstrap", "evil.test", "proxy.test", "secret"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("child environment retained %q: %s", forbidden, got)
		}
	}
	for _, required := range []string{"PATH=/usr/bin", "BB_TOKEN=actor", "HOME=/tmp/isolated"} {
		if !strings.Contains(got, required) {
			t.Fatalf("child environment missing %q: %s", required, got)
		}
	}
}

func TestMCPChildEnvProbe(t *testing.T) {
	path := os.Getenv("BBQA_ENV_PROBE_FILE")
	if path == "" {
		return
	}
	if err := os.WriteFile(path, []byte(strings.Join(os.Environ(), "\n")), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestMCPTransport_SpawnedChildReceivesOnlyAllowlistedEnvironment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "child-env.txt")
	env := filteredEnv([]string{"PATH=" + os.Getenv("PATH"), "HOME=/tmp/isolated", "BB_HOST=bitbucket.org", "BB_TOKEN=actor-token", "BB_CLOUD_BOOTSTRAP_TOKEN=bootstrap-secret", "BB_CLOUD_BASE_URL=http://redirect.test", "HTTPS_PROXY=http://proxy.test", "AWS_SECRET_ACCESS_KEY=aws-secret"})
	env = append(env, "BBQA_ENV_PROBE_FILE="+path)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	client, err := mcpclient.NewStdioMCPClientWithOptions(binary, env, []string{"-test.run=^TestMCPChildEnvProbe$"}, mcpCommandEnv(env))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	deadline := time.Now().Add(3 * time.Second)
	var content []byte
	for time.Now().Before(deadline) {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("helper process did not write its environment: %v", err)
	}
	got := string(content)
	for _, forbidden := range []string{"bootstrap-secret", "redirect.test", "proxy.test", "aws-secret"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("spawned MCP process inherited %q", forbidden)
		}
	}
	if !strings.Contains(got, "BB_TOKEN=actor-token") {
		t.Fatal("spawned MCP process did not receive the actor token")
	}
}

var _ mcptransport.StdioOption = mcpCommandEnv(nil)
