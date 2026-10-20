package acceptancecloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQueuePoll(t *testing.T) {
	t.Run("waits until independently observed", func(t *testing.T) {
		calls := 0
		err := queuePoll(context.Background(), time.Microsecond, func() (bool, error) { calls++; return calls == 3, nil })
		if err != nil || calls != 3 {
			t.Fatalf("err=%v calls=%d", err, calls)
		}
	})
	t.Run("fails on provider error", func(t *testing.T) {
		want := errors.New("HTTP 403")
		err := queuePoll(context.Background(), time.Microsecond, func() (bool, error) { return false, want })
		if !errors.Is(err, want) {
			t.Fatal(err)
		}
	})
	t.Run("missing queue cannot pass", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := queuePoll(ctx, time.Hour, func() (bool, error) { return false, nil })
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	})
}

func TestQueueHoldOpenRejectsMergeAfterCancellation(t *testing.T) {
	calls := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := queueHoldOpen(ctx, time.Microsecond, func() error {
		calls++
		if calls == 3 {
			return errors.New("PR must remain OPEN; got MERGED")
		}
		return nil
	})
	if err == nil || calls != 3 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
}

func TestQueueAcceptanceInvalidConfigCreatesNoManifest(t *testing.T) {
	dir := t.TempDir()
	report, err := RunMergeQueue(context.Background(), Config{ReportDir: dir}, io.Discard)
	if err == nil || report.Status == "pass" {
		t.Fatalf("report=%+v err=%v", report, err)
	}
	if _, err = os.Stat(filepath.Join(dir, "manifest.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("invalid configuration created a manifest")
	}
	b, err := os.ReadFile(filepath.Join(dir, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"skip_count": 0`) {
		t.Fatal("failure was recorded as skip")
	}
}

func TestQueueAcceptanceCleansOwnedRepoAfterSeedFailure(t *testing.T) {
	exists := false
	deleted := false
	uuid := "{fixture}"
	var marker string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer bootstrap" {
			t.Error("unexpected bootstrap credential")
		}
		switch {
		case strings.Contains(r.URL.Path, "/acceptance-sandbox/pipelines/"):
			_, _ = io.WriteString(w, `{"values":[]}`)
		case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/pullrequests"):
			_, _ = io.WriteString(w, `{"values":[]}`)
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/src"):
			w.WriteHeader(http.StatusForbidden)
		case r.Method == "POST":
			var body struct {
				Description string `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			marker = body.Description
			exists = true
			_, _ = fmt.Fprintf(w, `{"uuid":%q}`, uuid)
		case r.Method == "DELETE":
			deleted = true
			exists = false
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "GET":
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = fmt.Fprintf(w, `{"uuid":%q,"description":%q}`, uuid, marker)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{BootstrapToken: "bootstrap", ActorToken: "actor", Email: "fixture@example.com", Workspace: "bitbottle-qa-lab", Project: "BBQA", ReportDir: t.TempDir(), Binary: binary, BaseURL: server.URL}
	report, err := RunMergeQueue(context.Background(), cfg, io.Discard)
	if err == nil || !deleted || exists || report.Cleanup.Status != "pass" || report.Status == "pass" {
		t.Fatalf("err=%v deleted=%t exists=%t report=%+v", err, deleted, exists, report)
	}
	m, err := ReadManifest(report.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	if m.Resource.UUID != uuid {
		t.Fatal("cleanup identity not persisted")
	}
}

func TestQueueAcceptanceRejectsMissingPipelineScopeBeforeProvisioning(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/acceptance-sandbox/pipelines/") {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	cfg := Config{BootstrapToken: "bootstrap", ActorToken: "actor", Email: "fixture@example.com", Workspace: "bitbottle-qa-lab", Project: "BBQA", ReportDir: dir, Binary: binary, BaseURL: server.URL}
	report, err := RunMergeQueue(context.Background(), cfg, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "read:pipeline:bitbucket") || requests != 1 {
		t.Fatalf("unexpected preflight result: %v requests=%d", err, requests)
	}
	if report.Status != "infrastructure_failure" || report.Checks["bootstrap_pipeline_read"].Status != "fail" {
		t.Fatalf("invalid report: %+v", report)
	}
	if _, err = os.Stat(filepath.Join(dir, "manifest.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("scope failure created an ownership manifest")
	}
}
