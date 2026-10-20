package acceptancecloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestAutoMergeProbe_RecordsBothRoles404AndVerifiesCleanup(t *testing.T) {
	if internalProbeOrigin != "https://api.bitbucket.org" {
		t.Fatalf("internal probe origin = %q", internalProbeOrigin)
	}
	cfg, server, fixture := probeFixture(t, false)
	defer server.Close()
	report, err := runAutoMergeProbe(context.Background(), cfg, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if report.Outcome != "unconfirmed" || report.Cleanup.Status != "pass" || !fixture.deleted {
		t.Fatalf("report=%+v deleted=%v", report, fixture.deleted)
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
	for _, role := range []string{"actor", "bootstrap"} {
		for _, scheme := range []string{"bearer", "basic"} {
			prefix := role + "_internal_" + scheme
			mergeGet := report.Observations[prefix+"_merge_get"]
			if mergeGet.Status != http.StatusUnauthorized || mergeGet.Role != role || mergeGet.AuthScheme != scheme || mergeGet.CanCreatePostBuildMerge == nil || *mergeGet.CanCreatePostBuildMerge {
				t.Fatalf("internal merge GET observation=%+v", mergeGet)
			}
			if mergeGet.Method != http.MethodGet || !strings.HasSuffix(mergeGet.Path, "/post-build-merge") {
				t.Fatalf("unexpected internal GET request: %+v", mergeGet)
			}
			post := report.Observations[prefix+"_post_build_merge"]
			if post.Status != http.StatusUnauthorized || !strings.Contains(post.Message, "[redacted]") {
				t.Fatalf("internal POST observation=%+v", post)
			}
			if scheme == "basic" {
				token := cfg.ActorToken
				if role == "bootstrap" {
					token = cfg.BootstrapToken
				}
				encoded := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + token))
				if strings.Contains(post.Message, encoded) || strings.Contains(mergeGet.Message, encoded) {
					t.Fatal("returned observation exposed encoded Basic credentials")
				}
			}
			if report.Observations[prefix+"_pr_readback"].PRState != "OPEN" {
				t.Fatalf("internal POST did not verify OPEN PR for %s", prefix)
			}
		}
	}
	if !strings.Contains(report.Observations["actor_pr_get_before"].Message, "[redacted]") {
		t.Fatal("provider PR title was not sanitized in the returned report")
	}
	for _, key := range []string{"functional_rule_allow_auto_merge", "functional_rule_require_builds", "functional_build_inprogress", "functional_get_before", "functional_post_cancel", "functional_delete_cancel", "functional_cancel_build_success", "functional_final_build_inprogress", "functional_post_final", "functional_final_build_success", "functional_pr_final_1"} {
		if _, ok := report.Observations[key]; !ok {
			t.Errorf("missing functional observation %s", key)
		}
	}
	if report.Observations["functional_get_before"].PendingMergePresent == nil || *report.Observations["functional_get_before"].PendingMergePresent || report.Observations["functional_post_cancel"].PendingMergePresent == nil || !*report.Observations["functional_post_cancel"].PendingMergePresent {
		t.Fatalf("pending merge presence not recorded: GET=%+v POST=%+v", report.Observations["functional_get_before"], report.Observations["functional_post_cancel"])
	}
	if report.Observations["functional_pr_final_1"].PRState != "MERGED" {
		t.Fatalf("final merge observation=%+v", report.Observations["functional_pr_final_1"])
	}
	if len(report.Observations) > 80 {
		t.Fatalf("workflow summary limit would omit observations: %d", len(report.Observations))
	}
	for name := range report.Observations {
		if len(name) > 48 {
			t.Errorf("observation name exceeds workflow summary limit: %q", name)
		}
	}
	wantEvents := "build:INPROGRESS,post:1,delete,build:SUCCESSFUL,build:INPROGRESS,post:2,build:SUCCESSFUL"
	if strings.Join(fixture.events, ",") != wantEvents {
		t.Fatalf("functional sequence=%v", fixture.events)
	}
	if report.Observations["functional_get_before"].Path != "/internal/repositories/"+report.Repository+"/pullrequests/2/post-build-merge" {
		t.Fatalf("wrong functional route: %+v", report.Observations["functional_get_before"])
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
	for _, token := range []string{cfg.ActorToken, cfg.BootstrapToken} {
		encoded := base64.StdEncoding.EncodeToString([]byte(cfg.Email + ":" + token))
		if strings.Contains(string(b), encoded) {
			t.Fatal("probe persisted encoded Basic credentials")
		}
	}
	if _, err := ReadManifest(filepath.Join(cfg.ReportDir, "manifest.json")); err != nil {
		t.Fatal(err)
	}
}

func TestAutoMergeProbe_InvalidActorReadControlFailsAndCleans(t *testing.T) {
	cfg, server, fixture := probeFixture(t, true)
	defer server.Close()
	report, err := runAutoMergeProbe(context.Background(), cfg, server.URL)
	if err == nil || report.Outcome != "infrastructure_failure" || report.Cleanup.Status != "pass" || !fixture.deleted {
		t.Fatalf("report=%+v err=%v deleted=%v", report, err, fixture.deleted)
	}
	if _, ok := report.Observations["actor_auto_post"]; ok {
		t.Fatal("probed auto merge despite invalid actor control")
	}
}

func TestAutoMergeProbe_FunctionalRouteRejectionsStillCleanOwnedRepo(t *testing.T) {
	for _, mode := range []string{"post400", "delete403"} {
		t.Run(mode, func(t *testing.T) {
			cfg, server, fixture := probeFixture(t, false, mode)
			defer server.Close()
			report, err := runAutoMergeProbe(context.Background(), cfg, server.URL)
			if err != nil || report.Cleanup.Status != "pass" || !fixture.deleted || fixture.declines != 2 {
				t.Fatalf("report=%+v err=%v fixture=%+v", report, err, fixture)
			}
			if mode == "post400" && report.Observations["functional_post_cancel"].Status != http.StatusBadRequest {
				t.Fatalf("missing POST rejection observation: %+v", report.Observations["functional_post_cancel"])
			}
			if mode == "delete403" && report.Observations["functional_delete_cancel"].Status != http.StatusForbidden {
				t.Fatalf("missing DELETE rejection observation: %+v", report.Observations["functional_delete_cancel"])
			}
			if _, ok := report.Observations["functional_cancel_build_success"]; ok {
				t.Fatal("continued status mutation after rejected route")
			}
		})
	}
}

type probeFixtureState struct {
	deleted  bool
	events   []string
	declines int
}

func probeFixture(t *testing.T, denyActorRead bool, modes ...string) (Config, *httptest.Server, *probeFixtureState) {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-bitbottle")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho simulated-cli-error\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	mode := "ok"
	if len(modes) > 0 {
		mode = modes[0]
	}
	fixture := &probeFixtureState{}
	created := false
	marker := ""
	slug := ""
	type fixturePR struct{ title, description, state string }
	prs := map[int]fixturePR{}
	nextPR := 0
	functionalBuild := ""
	functionalPending := false
	functionalPosts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/internal/repositories/") {
			if !strings.HasPrefix(r.URL.Path, "/internal/repositories/bitbottle-qa-lab/"+slug+"/pullrequests/") || !strings.HasSuffix(r.URL.Path, "/post-build-merge") {
				t.Errorf("unexpected internal path %s", r.URL.Path)
			}
			pathParts := strings.Split(r.URL.Path, "/")
			prID := 0
			for i, part := range pathParts {
				if part == "pullrequests" && i+1 < len(pathParts) {
					prID, _ = strconv.Atoi(pathParts[i+1])
				}
			}
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Basic ") {
				email, token, ok := r.BasicAuth()
				if !ok || email != "actor@example.test" || (token != "actor-secret" && token != "bootstrap-secret") {
					t.Errorf("wrong basic authentication")
				}
			} else if auth != "Bearer actor-secret" && auth != "Bearer bootstrap-secret" {
				t.Errorf("wrong bearer authentication")
			}
			if prID == 1 && r.Method == http.MethodPost {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["close_source_branch"] != false || body["merge_strategy"] != "merge_commit" || body["newstatus"] != "fulfilled" || body["message"] != marker {
					t.Errorf("wrong post-build-merge payload: %v", body)
				}
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, `{"error":{"message":"Internal endpoint denied %s"}}`, auth)
				return
			}
			if prID == 2 {
				if r.Header.Get("Authorization") != "Bearer actor-secret" {
					t.Errorf("functional route must use the actor Bearer token")
				}
				switch r.Method {
				case http.MethodGet:
					if functionalPending {
						fmt.Fprint(w, `{"pending_merge":{"state":"INPROGRESS"}}`)
					} else {
						fmt.Fprint(w, `{"pending_merge":null}`)
					}
				case http.MethodPost:
					functionalPosts++
					fixture.events = append(fixture.events, fmt.Sprintf("post:%d", functionalPosts))
					if mode == "post400" {
						w.WriteHeader(http.StatusBadRequest)
						fmt.Fprint(w, `{"error":{"message":"destination branch does not support this diagnostic"}}`)
						return
					}
					functionalPending = true
					w.WriteHeader(http.StatusCreated)
					fmt.Fprint(w, `{"pending_merge":{"state":"INPROGRESS"}}`)
				case http.MethodDelete:
					if mode == "delete403" {
						w.WriteHeader(http.StatusForbidden)
						fmt.Fprint(w, `{"error":{"message":"delete is forbidden"}}`)
						return
					}
					fixture.events = append(fixture.events, "delete")
					functionalPending = false
					w.WriteHeader(http.StatusNoContent)
				}
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error":{"message":"Internal endpoint denied %s"},"can_create_post_build_merge":false}`, auth)
			return
		}
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
			if !created || fixture.deleted {
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
			fixture.deleted = true
			w.WriteHeader(204)
		case suffix == "/branch-restrictions" && r.Method == "POST":
			if !boot {
				w.WriteHeader(403)
				return
			}
			var rule map[string]any
			if err := json.NewDecoder(r.Body).Decode(&rule); err != nil || rule["branch_match_kind"] != "glob" || rule["pattern"] != "main" {
				t.Errorf("invalid functional branch rule: %v", rule)
			}
			kind := rule["kind"]
			if kind != "allow_auto_merge_when_builds_pass" && kind != "require_passing_builds_to_merge" {
				t.Errorf("unexpected functional branch rule: %v", rule)
			}
			if kind == "require_passing_builds_to_merge" && rule["value"] != float64(1) {
				t.Errorf("build requirement must be one successful build: %v", rule)
			}
			w.WriteHeader(http.StatusCreated)
		case strings.HasPrefix(suffix, "/commit/") && strings.HasSuffix(suffix, "/statuses/build") && r.Method == "POST":
			if !boot || !strings.HasPrefix(suffix, "/commit/source-hash/") {
				t.Errorf("functional build status must use bootstrap and fixture commit: %s", suffix)
			}
			var build struct {
				State string `json:"state"`
				Key   string `json:"key"`
				URL   string `json:"url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&build); err != nil || len(build.Key) != 16 || !strings.Contains(build.URL, "/pull-requests/2") {
				t.Errorf("invalid build status: %+v", build)
			}
			functionalBuild = build.State
			fixture.events = append(fixture.events, "build:"+build.State)
			w.WriteHeader(http.StatusCreated)
		case suffix == "/src" && r.Method == "POST":
			w.WriteHeader(201)
		case strings.HasPrefix(suffix, "/commit/") && r.Method == "GET":
			if strings.Contains(suffix, "auto-probe") {
				fmt.Fprint(w, `{"hash":"source-hash"}`)
			} else {
				fmt.Fprint(w, `{"hash":"abc"}`)
			}
		case suffix == "/refs/branches" && r.Method == "POST":
			var branch struct {
				Name   string `json:"name"`
				Target struct {
					Hash string `json:"hash"`
				} `json:"target"`
			}
			if err := json.NewDecoder(r.Body).Decode(&branch); err != nil || (!strings.HasPrefix(branch.Name, "auto-probe/") && !strings.HasPrefix(branch.Name, "acceptance/")) || branch.Target.Hash != "abc" {
				t.Errorf("invalid functional source branch: %+v", branch)
			}
			w.WriteHeader(201)
		case suffix == "/pullrequests" && r.Method == "POST":
			if !actor {
				w.WriteHeader(403)
				return
			}
			var p struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				Source      struct {
					Branch struct {
						Name string `json:"name"`
					} `json:"branch"`
				} `json:"source"`
				Destination struct {
					Branch struct {
						Name string `json:"name"`
					} `json:"branch"`
				} `json:"destination"`
			}
			_ = json.NewDecoder(r.Body).Decode(&p)
			if strings.HasPrefix(p.Title, "Functional") && (p.Description != marker+"\nfunctional diagnostic" || !strings.HasPrefix(p.Source.Branch.Name, "auto-probe/") || p.Destination.Branch.Name != "main") {
				t.Errorf("invalid functional PR: %+v", p)
			}
			nextPR++
			prs[nextPR] = fixturePR{title: p.Title, description: p.Description, state: "OPEN"}
			w.WriteHeader(201)
			fmt.Fprintf(w, `{"id":%d}`, nextPR)
		case strings.HasPrefix(suffix, "/pullrequests/") && r.Method == "GET":
			id, _ := strconv.Atoi(strings.TrimPrefix(suffix, "/pullrequests/"))
			pr := prs[id]
			if id == 2 && functionalPending && functionalBuild == "SUCCESSFUL" && functionalPosts >= 2 {
				pr.state = "MERGED"
				prs[id] = pr
			}
			readTitle := pr.title
			if id == 1 && !strings.HasSuffix(pr.title, " control") {
				readTitle += " actor@example.test"
			}
			pending := "null"
			if id == 2 && functionalPending {
				pending = `{"state":"INPROGRESS"}`
			}
			fmt.Fprintf(w, `{"id":%d,"title":%q,"description":%q,"state":%q,"pending_merge":%s}`, id, readTitle, pr.description, pr.state, pending)
		case suffix == "/pullrequests/1" && r.Method == "PUT":
			if !actor {
				w.WriteHeader(403)
				return
			}
			var p struct {
				Title string `json:"title"`
			}
			_ = json.NewDecoder(r.Body).Decode(&p)
			pr := prs[1]
			pr.title = p.Title
			prs[1] = pr
			fmt.Fprintf(w, `{"id":1,"title":%q}`, pr.title)
		case suffix == "/pullrequests/1/auto-merge":
			w.WriteHeader(404)
			fmt.Fprint(w, `{"error":{"message":"No auto-merge endpoint for actor@example.test"}}`)
		case suffix == "/pullrequests" && r.Method == "GET":
			fmt.Fprint(w, `{"values":[`)
			first := true
			for id, pr := range prs {
				if !first {
					fmt.Fprint(w, ",")
				}
				first = false
				fmt.Fprintf(w, `{"id":%d,"state":%q}`, id, pr.state)
			}
			fmt.Fprint(w, `]}`)
		case strings.HasPrefix(suffix, "/pullrequests/") && strings.HasSuffix(suffix, "/decline") && r.Method == "POST":
			idText := strings.TrimSuffix(strings.TrimPrefix(suffix, "/pullrequests/"), "/decline")
			id, _ := strconv.Atoi(idText)
			pr := prs[id]
			pr.state = "DECLINED"
			prs[id] = pr
			fixture.declines++
			w.WriteHeader(200)
		default:
			http.Error(w, "unexpected probe route", 500)
		}
	}))
	cfg := Config{BootstrapToken: "bootstrap-secret", ActorToken: "actor-secret", Email: "actor@example.test", Workspace: "bitbottle-qa-lab", Project: "BBQA", ReportDir: dir, Binary: script, BaseURL: server.URL}
	return cfg, server, fixture
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

func TestProbeInternalRequest_AuthAndRedirectBoundary(t *testing.T) {
	redirected := false
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { redirected = true }))
	defer target.Close()
	for _, scheme := range []string{"bearer", "basic"} {
		t.Run(scheme, func(t *testing.T) {
			origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/internal/repositories/workspace/repo/pullrequests/1/post-build-merge" || r.Method != http.MethodGet || r.Header.Get("Cookie") != "" {
					t.Errorf("unexpected request path or cookie")
				}
				if scheme == "basic" {
					email, token, ok := r.BasicAuth()
					if !ok || email != "actor@example.test" || token != "actor-secret" {
						t.Errorf("incorrect basic auth")
					}
				} else if r.Header.Get("Authorization") != "Bearer actor-secret" {
					t.Errorf("incorrect bearer auth")
				}
				http.Redirect(w, r, target.URL, http.StatusFound)
			}))
			defer origin.Close()
			cfg := Config{ActorToken: "actor-secret", Email: "actor@example.test"}
			got := probeInternalRequest(context.Background(), internalProbeHTTPClient(), origin.URL, "/internal/repositories/workspace/repo/pullrequests/1/post-build-merge", http.MethodGet, nil, "actor", scheme, cfg.Email, cfg.ActorToken, cfg)
			if got.Status != http.StatusFound || got.Role != "actor" || got.AuthScheme != scheme || redirected {
				t.Fatalf("redirect escaped boundary: %+v, redirected=%v", got, redirected)
			}
		})
	}
}
