package script_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

var update = flag.Bool("update", false, "update golden files")

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:           "testdata",
		Setup:         setup,
		Cmds:          customCmds(),
		UpdateScripts: *update,
	})
	// upgrade sub-corpus lives in testdata/upgrade/ so the filter
	// -run TestScript/upgrade matches all four migration-path scripts.
	t.Run("upgrade", func(t *testing.T) {
		testscript.Run(t, testscript.Params{
			Dir:           "testdata/upgrade",
			Setup:         setup,
			Cmds:          customCmds(),
			UpdateScripts: *update,
		})
	})
}

// setup runs before each script: scrubs env and provides a hermetic HOME.
func setup(env *testscript.Env) error {
	// Scrub env vars that could bleed state between tests.
	// BB_ prefix covers BB_TOKEN, BB_HOST, BB_CLOUD_BASE_URL, BB_CONFIG_DIR, etc.
	scrubPrefixes := []string{"BB_", "BITBOTTLE_", "HTTPS_PROXY", "GIT_"}
	scrubExact := []string{
		"XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME",
		"NO_COLOR", "HOME",
	}
	for i, kv := range env.Vars {
		k, _, _ := strings.Cut(kv, "=")
		for _, pfx := range scrubPrefixes {
			if strings.HasPrefix(k, pfx) {
				env.Vars[i] = k + "="
			}
		}
		for _, exact := range scrubExact {
			if k == exact {
				env.Vars[i] = k + "="
			}
		}
	}

	// Point HOME at the per-test work dir so config + keyring are hermetic.
	env.Vars = append(env.Vars, "HOME="+env.WorkDir)
	// Use file-based keyring so tests never touch the OS keyring daemon.
	env.Vars = append(env.Vars, "BITBOTTLE_ALLOW_INSECURE_STORE=1")
	// Disable interactive prompts.
	env.Vars = append(env.Vars, "BB_PROMPT_DISABLED=1")
	// Point config dir at a subdir of the test work dir.
	cfgDir := env.WorkDir + "/.config/bitbottle"
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return err
	}
	env.Vars = append(env.Vars, "BB_CONFIG_DIR="+cfgDir)
	// Export the config dir path so scripts can reference it.
	env.Setenv("BB_CONFIG_DIR", cfgDir)
	return nil
}

// customCmds returns the extra testscript commands available in scripts.
func customCmds() map[string]func(ts *testscript.TestScript, neg bool, args []string) {
	return map[string]func(ts *testscript.TestScript, neg bool, args []string){
		"bb-fake": bbFakeCmd,
	}
}

// bbFakeCmd implements the bb-fake testscript command.
//
// Usage:
//
//	bb-fake server   — start a fake Bitbucket Server (DC) and set env:
//	                     BB_FAKE_SERVER_URL, BB_TOKEN, BB_HOST
//	bb-fake cloud    — start a fake Bitbucket Cloud and set env:
//	                     BB_FAKE_CLOUD_URL, BB_TOKEN, BB_HOST
//
// The server writes its URL into the script env and also writes a minimal
// hosts.yml so bitbottle can resolve the host without interactive login.
func bbFakeCmd(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) < 1 {
		ts.Fatalf("bb-fake: missing subcommand (server|cloud)")
	}
	sub := args[0]
	var srv *httptest.Server
	var hostname string
	const fakeToken = "fake-test-token"
	const fakeUser = "testuser"

	cfgDir := ts.Getenv("BB_CONFIG_DIR")
	if cfgDir == "" {
		ts.Fatalf("bb-fake: BB_CONFIG_DIR not set")
	}

	switch sub {
	case "server":
		srv = buildServerStubs(ts)
		hostname = extractHost(srv.URL)
		ts.Setenv("BB_FAKE_SERVER_URL", srv.URL)
		ts.Setenv("BB_HOST", hostname)
		ts.Setenv("BB_TOKEN", fakeToken)
		writeHostsYML(ts, cfgDir, hostname, fakeUser, fakeToken, "server")
	case "cloud":
		srv = buildCloudStubs(ts)
		hostname = "bitbucket.org"
		ts.Setenv("BB_FAKE_CLOUD_URL", srv.URL)
		ts.Setenv("BB_HOST", hostname)
		ts.Setenv("BB_TOKEN", fakeToken)
		writeHostsYML(ts, cfgDir, hostname, fakeUser, fakeToken, "cloud")
		// Redirect all Cloud API calls to the fake server.
		ts.Setenv("BB_CLOUD_BASE_URL", srv.URL+"/2.0")
	default:
		ts.Fatalf("bb-fake: unknown subcommand %q; want server|cloud", sub)
	}

	// Propagate the fake server's TLS client cert so bitbottle accepts it.
	// We pass --skip-tls-verify in the scripts instead to keep it simple.
	ts.Setenv("BB_FAKE_TLS_URL", srv.URL)
}

// extractHost returns host:port from an https URL.
func extractHost(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimSuffix(u, "/")
	return u
}

// writeHostsYML writes a minimal hosts.yml so bitbottle treats the fake as
// authenticated. The token goes inline (pre-migrate format) so the tests
// don't need to touch the OS keyring. oauth_token in hosts.yml is valid for
// BITBOTTLE_ALLOW_INSECURE_STORE=1 contexts.
func writeHostsYML(ts *testscript.TestScript, cfgDir, hostname, user, token, backendType string) {
	content := fmt.Sprintf("%s:\n  user: %s\n  oauth_token: %s\n  git_protocol: https\n  backend_type: %s\n  skip_tls_verify: true\n",
		hostname, user, token, backendType)
	if err := os.WriteFile(cfgDir+"/hosts.yml", []byte(content), 0o600); err != nil {
		ts.Fatalf("bb-fake: writing hosts.yml: %v", err)
	}
}

// buildServerStubs constructs a TLS httptest.Server with Bitbucket Server
// fixture responses.
func buildServerStubs(ts *testscript.TestScript) *httptest.Server {
	stubs := []testhelpers.StubResponse{
		// application-properties — version probe called on first request
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/application-properties",
			Status:     http.StatusOK,
			Body: map[string]any{
				"version":     "8.0.0",
				"buildNumber": "8000000",
				"displayName": "Bitbucket",
			},
		},
		// whoami — used by auth status
		{
			Method:     http.MethodGet,
			PathSuffix: "/plugins/servlet/applinks/whoami",
			Status:     http.StatusOK,
			Body:       "testuser",
		},
		// user profile
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/users/testuser",
			Status:     http.StatusOK,
			Body:       map[string]any{"slug": "testuser", "displayName": "Test User", "emailAddress": "test@example.com"},
		},
		// repo list
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/repos",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{"slug": "alpha-repo", "project": map[string]any{"key": "PROJ"}},
				map[string]any{"slug": "beta-repo", "project": map[string]any{"key": "PROJ"}},
			}),
		},
		// PR list
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/alpha-repo/pull-requests",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				serverPR(1, "Fix the bug", "OPEN"),
				serverPR(2, "Add feature", "OPEN"),
			}),
		},
		// PR view
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/alpha-repo/pull-requests/1",
			Status:     http.StatusOK,
			Body:       serverPR(1, "Fix the bug", "OPEN"),
		},
		// 404 — for errfmt_not_found
		{
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/missing-repo/pull-requests",
			Status:     http.StatusNotFound,
			Body:       map[string]any{"errors": []any{map[string]any{"message": "Repository not found"}}},
		},
		// 401 — for errfmt_auth
		{
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/bad-auth/pull-requests",
			Status:     http.StatusUnauthorized,
			Body:       map[string]any{"errors": []any{map[string]any{"message": "Unauthorized"}}},
		},
		// REST API root — used by auth doctor reachability probe
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0",
			Status:     http.StatusOK,
			Body:       map[string]any{},
		},
		// PUT PR #1 — used by pr unready (GET-then-PUT) and pr edit --title
		{
			Method:     http.MethodPut,
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/alpha-repo/pull-requests/1",
			Status:     http.StatusOK,
			Body:       serverPR(1, "Fix the bug", "OPEN"),
		},
		// DELETE PR participant — used by pr edit --remove-reviewer
		{
			Method:     http.MethodDelete,
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/alpha-repo/pull-requests/1/participants/alice",
			Status:     http.StatusNoContent,
		},
		// GET reviewer-group conditions — used by pr reviewer-group list
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/default-reviewers/1.0/projects/PROJ/repos/alpha-repo/conditions",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{
					"id":                1,
					"requiredApprovals": 1,
					"reviewers":         []any{map[string]any{"slug": "bob", "displayName": "Bob"}},
					"sourceMatcher":     map[string]any{"id": "team-reviews", "displayId": "team-reviews"},
					"targetMatcher":     map[string]any{"id": "ANY_REF_MATCHER_ID", "displayId": "Any branch"},
				},
			}),
		},
		// POST suggestion apply — used by pr suggestion apply
		{
			Method:     http.MethodPost,
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/alpha-repo/pull-requests/1/comments/5/suggestions/3/apply",
			Status:     http.StatusOK,
			Body:       map[string]any{"commitHash": "deadbeef1234", "commitMessage": "Apply suggestion"},
		},
		// GET PR activity — used by pr activity --json
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/alpha-repo/pull-requests/1/activities",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{
					"action":      "APPROVED",
					"createdDate": int64(1714550400000),
					"user":        map[string]any{"slug": "alice", "displayName": "Alice"},
				},
				map[string]any{
					"action":      "COMMENTED",
					"createdDate": int64(1714554000000),
					"user":        map[string]any{"slug": "bob", "displayName": "Bob"},
				},
			}),
		},
		// GET pr-settings — used by repo pr-settings get and by set (GET-then-POST merge)
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/alpha-repo/settings/pull-requests",
			Status:     http.StatusOK,
			Body: map[string]any{
				"requiredApprovers":        2,
				"requiredAllApprovers":     false,
				"requiredAllTasksComplete": false,
				"requiredSuccessfulBuilds": 1,
				"mergeConfig": map[string]any{
					"defaultStrategy": map[string]any{"id": "no-ff"},
					"strategies":      []any{map[string]any{"id": "no-ff"}, map[string]any{"id": "squash"}},
				},
			},
		},
		// POST pr-settings — used by repo pr-settings set
		{
			Method:     http.MethodPost,
			PathSuffix: "/rest/api/1.0/projects/PROJ/repos/alpha-repo/settings/pull-requests",
			Status:     http.StatusOK,
			Body: map[string]any{
				"requiredApprovers":        3,
				"requiredAllApprovers":     false,
				"requiredAllTasksComplete": false,
				"requiredSuccessfulBuilds": 1,
				"mergeConfig": map[string]any{
					"defaultStrategy": map[string]any{"id": "no-ff"},
					"strategies":      []any{map[string]any{"id": "no-ff"}, map[string]any{"id": "squash"}},
				},
			},
		},
		// GET admin groups — used by group list
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/admin/groups",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{"name": "developers"},
				map[string]any{"name": "admins"},
			}),
		},
		// POST admin groups — used by group create
		{
			Method:     http.MethodPost,
			PathSuffix: "/rest/api/1.0/admin/groups",
			Status:     http.StatusOK,
			Body:       map[string]any{"name": "newgroup"},
		},
		// DELETE admin groups — used by group delete
		{
			Method:     http.MethodDelete,
			PathSuffix: "/rest/api/1.0/admin/groups",
			Status:     http.StatusNoContent,
		},
		// GET admin groups more-members — used by group member list
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/admin/groups/more-members",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{"name": "alice", "displayName": "Alice", "emailAddress": "alice@example.com"},
				map[string]any{"name": "bob", "displayName": "Bob", "emailAddress": "bob@example.com"},
			}),
		},
		// POST admin users add-group — used by group member add
		{
			Method:     http.MethodPost,
			PathSuffix: "/rest/api/1.0/admin/users/add-group",
			Status:     http.StatusNoContent,
		},
		// POST admin users remove-group — used by group member remove
		{
			Method:     http.MethodPost,
			PathSuffix: "/rest/api/1.0/admin/users/remove-group",
			Status:     http.StatusNoContent,
		},
		// GET projects list — used by project server-list
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/projects",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{"key": "PRJ", "name": "My Project", "description": "A project", "public": false,
					"links": map[string]any{"self": []any{map[string]any{"href": "https://example.com/projects/PRJ"}}}},
				map[string]any{"key": "DEV", "name": "Dev Project", "description": "", "public": true,
					"links": map[string]any{"self": []any{map[string]any{"href": "https://example.com/projects/DEV"}}}},
			}),
		},
		// GET project view — used by project view
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/api/1.0/projects/PRJ",
			Status:     http.StatusOK,
			Body: map[string]any{
				"key": "PRJ", "name": "My Project", "description": "A project", "public": false,
				"links": map[string]any{"self": []any{map[string]any{"href": "https://example.com/projects/PRJ"}}},
			},
		},
		// POST project create — used by project create
		{
			Method:     http.MethodPost,
			PathSuffix: "/rest/api/1.0/projects",
			Status:     http.StatusCreated,
			Body: map[string]any{
				"key": "NEWPRJ", "name": "New Project", "description": "", "public": false,
				"links": map[string]any{"self": []any{map[string]any{"href": "https://example.com/projects/NEWPRJ"}}},
			},
		},
		// PUT project edit — used by project edit (GET-then-PUT)
		{
			Method:     http.MethodPut,
			PathSuffix: "/rest/api/1.0/projects/PRJ",
			Status:     http.StatusOK,
			Body: map[string]any{
				"key": "PRJ", "name": "Updated Project", "description": "A project", "public": false,
				"links": map[string]any{"self": []any{map[string]any{"href": "https://example.com/projects/PRJ"}}},
			},
		},
		// DELETE project — used by project delete
		{
			Method:     http.MethodDelete,
			PathSuffix: "/rest/api/1.0/projects/DELPRJ",
			Status:     http.StatusNoContent,
		},
		// GET PAT list — used by auth pat list
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/access-tokens/1.0/users/testuser",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{
					"id":          "1",
					"name":        "CI Token",
					"permissions": []any{"REPO_READ"},
					"createdDate": int64(1716556800000),
				},
			}),
		},
		// PUT PAT create — used by auth pat create
		{
			Method:     http.MethodPut,
			PathSuffix: "/rest/access-tokens/1.0/users/testuser",
			Status:     http.StatusOK,
			Body: map[string]any{
				"id":          "2",
				"name":        "CI Token",
				"permissions": []any{"REPO_READ", "REPO_WRITE"},
				"createdDate": int64(1716556800000),
				"token":       "BBDC-generatedtoken",
			},
		},
		// DELETE PAT revoke — used by auth pat revoke
		{
			Method:     http.MethodDelete,
			PathSuffix: "/rest/access-tokens/1.0/users/testuser/1",
			Status:     http.StatusNoContent,
		},
		// SSH key list — used by ssh-key list (Server/DC)
		{
			Method:     http.MethodGet,
			PathSuffix: "/rest/ssh/1.0/keys",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{"id": 1, "label": "Laptop", "text": "ssh-rsa AAAA...laptop"},
				map[string]any{"id": 2, "label": "Desktop", "text": "ssh-rsa AAAA...desktop"},
			}),
		},
		// compare/commits (ahead) — used by branch compare (Server/DC)
		{
			Method:     http.MethodGet,
			PathSuffix: "/compare/commits",
			Status:     http.StatusOK,
			Body: testhelpers.PagedResponse([]any{
				map[string]any{"id": "abc1234", "message": "feat: new thing\n", "author": map[string]any{"name": "Test User"}, "authorTimestamp": int64(1704067200000)},
			}),
		},
	}
	return newTLSServer(ts, stubs...)
}

// buildCloudStubs constructs a TLS httptest.Server with Bitbucket Cloud
// fixture responses.
func buildCloudStubs(ts *testscript.TestScript) *httptest.Server {
	stubs := []testhelpers.StubResponse{
		// current user
		{
			Method:     http.MethodGet,
			PathSuffix: "/user",
			Status:     http.StatusOK,
			Body:       map[string]any{"uuid": "{uuid}", "username": "testuser", "display_name": "Test User"},
		},
		// repo list
		{
			Method:     http.MethodGet,
			PathSuffix: "/repositories/testworkspace",
			Status:     http.StatusOK,
			Body: testhelpers.CloudPagedResponse([]any{
				map[string]any{"full_name": "testworkspace/cloud-repo-a", "scm": "git", "is_private": false},
				map[string]any{"full_name": "testworkspace/cloud-repo-b", "scm": "git", "is_private": true},
			}),
		},
		// PR list
		{
			Method:     http.MethodGet,
			PathSuffix: "/repositories/testworkspace/cloud-repo-a/pullrequests",
			Status:     http.StatusOK,
			Body: testhelpers.CloudPagedResponse([]any{
				cloudPR(10, "Cloud fix", "OPEN"),
				cloudPR(11, "Cloud feature", "OPEN"),
			}),
		},
		// PR view
		{
			Method:     http.MethodGet,
			PathSuffix: "/repositories/testworkspace/cloud-repo-a/pullrequests/10",
			Status:     http.StatusOK,
			Body:       cloudPR(10, "Cloud fix", "OPEN"),
		},
		// 404 — for errfmt_not_found
		{
			PathSuffix: "/repositories/testworkspace/no-such-repo/pullrequests",
			Status:     http.StatusNotFound,
			Body:       map[string]any{"type": "error", "error": map[string]any{"message": "Repository not found"}},
		},
		// 401 — for errfmt_auth
		{
			PathSuffix: "/repositories/testworkspace/bad-auth/pullrequests",
			Status:     http.StatusUnauthorized,
			Body:       map[string]any{"type": "error", "error": map[string]any{"message": "Unauthorized"}},
		},
		// pipelines_config GET — used by pipeline config get
		{
			Method:     http.MethodGet,
			PathSuffix: "/repositories/testworkspace/cloud-repo-a/pipelines_config",
			Status:     http.StatusOK,
			Body:       map[string]any{"type": "repository_pipeline_settings", "enabled": true},
		},
		// pipelines_config PUT — used by pipeline config enable/disable
		{
			Method:     http.MethodPut,
			PathSuffix: "/repositories/testworkspace/cloud-repo-a/pipelines_config",
			Status:     http.StatusOK,
			Body:       map[string]any{"type": "repository_pipeline_settings", "enabled": false},
		},
		// SSH key list — used by ssh-key list (Cloud)
		{
			Method:     http.MethodGet,
			PathSuffix: "/users/testuser/ssh-keys",
			Status:     http.StatusOK,
			Body: testhelpers.CloudPagedResponse([]any{
				map[string]any{"id": 1, "label": "Laptop", "key": "ssh-rsa AAAA...laptop"},
			}),
		},
		// test report summary — used by pipeline test-report view
		{
			Method:     http.MethodGet,
			PathSuffix: "/test_reports",
			Status:     http.StatusOK,
			Body: map[string]any{
				"total_count":         10,
				"success_count":       8,
				"failed_count":        2,
				"error_count":         0,
				"skipped_count":       1,
				"duration_in_seconds": 12.5,
			},
		},
		// test cases list — used by pipeline test-case list
		{
			Method:     http.MethodGet,
			PathSuffix: "/test_reports/test_cases",
			Status:     http.StatusOK,
			Body: testhelpers.CloudPagedResponse([]any{
				map[string]any{"name": "TestFoo", "class_name": "com.example.Foo", "status": "FAILED", "duration_in_seconds": 0.5, "error_details": "assertion failed"},
				map[string]any{"name": "TestBar", "class_name": "com.example.Bar", "status": "PASSED", "duration_in_seconds": 0.2},
			}),
		},
		// branch compare commits (ahead) — used by branch compare (Cloud)
		{
			Method:     http.MethodGet,
			PathSuffix: "/commits/feature",
			Status:     http.StatusOK,
			Body: testhelpers.CloudPagedResponse([]any{
				map[string]any{"hash": "abc1234", "message": "feat: new thing\n", "author": map[string]any{"raw": "Test User"}, "date": "2024-01-01T00:00:00Z", "links": map[string]any{"html": map[string]any{"href": ""}}},
			}),
		},
		// branch compare commits (behind) — used by branch compare (Cloud)
		{
			Method:     http.MethodGet,
			PathSuffix: "/commits/main",
			Status:     http.StatusOK,
			Body:       testhelpers.CloudPagedResponse([]any{}),
		},
	}
	return newTLSServer(ts, stubs...)
}

// newTLSServer starts a TLS server dispatching to the given stubs.
// Unmatched requests call ts.Fatalf.
func newTLSServer(ts *testscript.TestScript, stubs ...testhelpers.StubResponse) *httptest.Server {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, s := range stubs {
			if s.Method != "" && !strings.EqualFold(s.Method, r.Method) {
				continue
			}
			if s.PathSuffix != "" && !strings.HasSuffix(r.URL.Path, s.PathSuffix) {
				continue
			}
			if s.Handler != nil {
				s.Handler(w, r)
				return
			}
			status := s.Status
			if status == 0 {
				status = http.StatusOK
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			if s.Body != nil {
				if err := json.NewEncoder(w).Encode(s.Body); err != nil {
					ts.Fatalf("newTLSServer: encode stub: %v", err)
				}
			}
			return
		}
		ts.Fatalf("newTLSServer: unmatched request %s %s", r.Method, r.URL.Path)
		http.Error(w, "unmatched", http.StatusNotImplemented)
	}))
	ts.Defer(srv.Close)
	return srv
}

func serverPR(id int, title, state string) map[string]any {
	return map[string]any{
		"id":    id,
		"title": title,
		"state": state,
		"author": map[string]any{
			"user": map[string]any{"slug": "alice", "displayName": "Alice"},
		},
		"fromRef": map[string]any{"displayId": "feature/fix", "id": "refs/heads/feature/fix"},
		"toRef":   map[string]any{"displayId": "main", "id": "refs/heads/main"},
		"links":   map[string]any{"self": []any{map[string]any{"href": "https://bitbucket.example.com/pr/" + fmt.Sprint(id)}}},
		"version": 0,
	}
}

func cloudPR(id int, title, state string) map[string]any {
	return map[string]any{
		"id":    id,
		"title": title,
		"state": state,
		"author": map[string]any{
			"display_name": "Alice",
			"uuid":         "{alice-uuid}",
		},
		"source":      map[string]any{"branch": map[string]any{"name": "feature/fix"}},
		"destination": map[string]any{"branch": map[string]any{"name": "main"}},
		"links":       map[string]any{"html": map[string]any{"href": "https://bitbucket.org/pr/" + fmt.Sprint(id)}},
	}
}
