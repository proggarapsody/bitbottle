package acceptancecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutoMergeProbe_RecordsBothRoles404AndVerifiesCleanup(t *testing.T) {
	cfg, server, deleted := probeFixture(t, false)
	defer server.Close()
	report, err := RunAutoMergeProbe(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if report.Outcome != "unconfirmed" || report.Cleanup.Status != "pass" || !*deleted {
		t.Fatalf("report=%+v deleted=%v", report, *deleted)
	}
	for _, key := range []string{"actor_pr_get_before", "actor_pr_write", "actor_pr_get_after_write", "cli_auto", "cli_auto_off", "actor_auto_post", "actor_auto_get", "actor_auto_delete", "bootstrap_auto_post_control", "bootstrap_auto_delete_control", "bootstrap_auto_delete_pr_readback"} {
		if _, ok := report.Observations[key]; !ok {
			t.Errorf("missing observation %s", key)
		}
	}
	if report.Observations["actor_auto_post"].Status != 404 || report.Observations["bootstrap_auto_post_control"].Status != 404 {
		t.Fatal("raw 404 evidence absent")
	}
	if report.Observations["bootstrap_auto_delete_pr_readback"].PRState != "OPEN" {
		t.Fatal("bootstrap DELETE did not verify the PR stayed open")
	}
	if !strings.Contains(report.Observations["actor_pr_get_before"].Message, "[redacted]") {
		t.Fatal("provider PR title was not sanitized in the returned report")
	}
	b, err := os.ReadFile(filepath.Join(cfg.ReportDir, "probe.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{cfg.ActorToken, cfg.BootstrapToken, cfg.Email} {
		if strings.Contains(string(b), secret) {
			t.Fatalf("probe persisted secret %q", secret)
		}
	}
	if _, err := ReadManifest(filepath.Join(cfg.ReportDir, "manifest.json")); err != nil {
		t.Fatal(err)
	}
}

func TestAutoMergeProbe_InvalidActorReadControlFailsAndCleans(t *testing.T) {
	cfg, server, deleted := probeFixture(t, true)
	defer server.Close()
	report, err := RunAutoMergeProbe(context.Background(), cfg)
	if err == nil || report.Outcome != "infrastructure_failure" || report.Cleanup.Status != "pass" || !*deleted {
		t.Fatalf("report=%+v err=%v deleted=%v", report, err, *deleted)
	}
	if _, ok := report.Observations["actor_auto_post"]; ok {
		t.Fatal("probed auto merge despite invalid actor control")
	}
}

func probeFixture(t *testing.T, denyActorRead bool) (Config, *httptest.Server, *bool) {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-bitbottle")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho simulated-cli-error\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	deleted := false
	created := false
	marker := ""
	slug := ""
	title := ""
	prState := "OPEN"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.HasPrefix(r.URL.Path, "/repositories/bitbottle-qa-lab/") {
			http.Error(w, "bad path", 400)
			return
		}
		rest := strings.TrimPrefix(r.URL.Path, "/repositories/bitbottle-qa-lab/")
		parts := strings.SplitN(rest, "/", 2)
		if slug == "" {
			slug = parts[0]
		}
		if slug != parts[0] {
			http.Error(w, "wrong repo", 400)
			return
		}
		suffix := ""
		if len(parts) == 2 {
			suffix = "/" + parts[1]
		}
		actor := r.Header.Get("Authorization") == "Bearer actor-secret"
		boot := r.Header.Get("Authorization") == "Bearer bootstrap-secret"
		if !actor && !boot {
			w.WriteHeader(401)
			return
		}
		if denyActorRead && actor && r.Method == "GET" && strings.HasPrefix(suffix, "/pullrequests/1") && !strings.HasSuffix(suffix, "/auto-merge") {
			w.WriteHeader(401)
			return
		}
		switch {
		case suffix == "" && r.Method == "GET":
			if !created || deleted {
				w.WriteHeader(404)
				return
			}
			fmt.Fprintf(w, `{"uuid":"repo-uuid","description":%q}`, marker)
		case suffix == "" && r.Method == "POST":
			if !boot {
				w.WriteHeader(403)
				return
			}
			var p struct {
				Description string `json:"description"`
			}
			_ = json.NewDecoder(r.Body).Decode(&p)
			marker = p.Description
			created = true
			w.WriteHeader(201)
			fmt.Fprint(w, `{"uuid":"repo-uuid"}`)
		case suffix == "" && r.Method == "DELETE":
			if !boot {
				w.WriteHeader(403)
				return
			}
			deleted = true
			w.WriteHeader(204)
		case suffix == "/src" && r.Method == "POST":
			w.WriteHeader(201)
		case suffix == "/commit/main" && r.Method == "GET":
			fmt.Fprint(w, `{"hash":"abc"}`)
		case suffix == "/refs/branches" && r.Method == "POST":
			w.WriteHeader(201)
		case suffix == "/pullrequests" && r.Method == "POST":
			if !actor {
				w.WriteHeader(403)
				return
			}
			var p struct {
				Title string `json:"title"`
			}
			_ = json.NewDecoder(r.Body).Decode(&p)
			title = p.Title
			w.WriteHeader(201)
			fmt.Fprint(w, `{"id":1}`)
		case suffix == "/pullrequests/1" && r.Method == "GET":
			readTitle := title
			if !strings.HasSuffix(title, " control") {
				readTitle += " actor@example.test"
			}
			fmt.Fprintf(w, `{"id":1,"title":%q,"description":%q,"state":%q}`, readTitle, marker, prState)
		case suffix == "/pullrequests/1" && r.Method == "PUT":
			if !actor {
				w.WriteHeader(403)
				return
			}
			var p struct {
				Title string `json:"title"`
			}
			_ = json.NewDecoder(r.Body).Decode(&p)
			title = p.Title
			fmt.Fprintf(w, `{"id":1,"title":%q}`, title)
		case suffix == "/pullrequests/1/auto-merge":
			w.WriteHeader(404)
			fmt.Fprint(w, `{"error":{"message":"No auto-merge endpoint for actor@example.test"}}`)
		case suffix == "/pullrequests" && r.Method == "GET":
			fmt.Fprintf(w, `{"values":[{"id":1,"state":%q}]}`, prState)
		case suffix == "/pullrequests/1/decline" && r.Method == "POST":
			prState = "DECLINED"
			w.WriteHeader(200)
		default:
			http.Error(w, "unexpected probe route", 500)
		}
	}))
	cfg := Config{BootstrapToken: "bootstrap-secret", ActorToken: "actor-secret", Email: "actor@example.test", Workspace: "bitbottle-qa-lab", Project: "BBQA", ReportDir: dir, Binary: script, BaseURL: server.URL}
	return cfg, server, &deleted
}

func TestProbeReadPR_SeparatesPresenceFromExplicitEnabled(t *testing.T) {
	for _, tc := range []struct {
		name, field      string
		present, enabled *bool
		strategy, shape  string
	}{
		{name: "absent", field: "", present: boolPtr(false)},
		{name: "strategy only", field: `,"auto_merge":{"merge_strategy":"merge_commit"}`, present: boolPtr(true), strategy: "merge_commit"},
		{name: "explicit disabled", field: `,"auto_merge":{"enabled":false}`, present: boolPtr(true), enabled: boolPtr(false)},
		{name: "unknown shape", field: `,"auto_merge":["unexpected"]`, present: boolPtr(true), shape: "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{"id":1,"title":"sample","state":"OPEN"%s}`, tc.field)
			}))
			defer server.Close()
			client, err := newAPI(server.URL, "fixture")
			if err != nil {
				t.Fatal(err)
			}
			got, err := probeReadPR(context.Background(), client, "/pullrequests/1", 1)
			if err != nil {
				t.Fatal(err)
			}
			if got.AutoMergePresent == nil || *got.AutoMergePresent != *tc.present {
				t.Fatalf("presence=%v", got.AutoMergePresent)
			}
			if (got.AutoMergeEnabled == nil) != (tc.enabled == nil) {
				t.Fatalf("enabled=%v", got.AutoMergeEnabled)
			}
			if tc.enabled != nil && *got.AutoMergeEnabled != *tc.enabled {
				t.Fatalf("enabled=%v", *got.AutoMergeEnabled)
			}
			if got.AutoMergeStrategy != tc.strategy || got.AutoMergeShape != tc.shape {
				t.Fatalf("observation=%+v", got)
			}
		})
	}
}

func boolPtr(v bool) *bool { return &v }
