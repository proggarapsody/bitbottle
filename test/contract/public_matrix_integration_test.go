package contract_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
	"github.com/proggarapsody/bitbottle/pkg/errfmt"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
)

type matrixProvider struct {
	name, host, apiRoot, prPath string
}

var matrixProviders = []matrixProvider{
	{"cloud", "bitbucket.org", "/2.0", "/repositories/TEAM/service/pullrequests"},
	{"server", "git.example.test", "/rest/api/1.0", "/projects/TEAM/repos/service/pull-requests"},
}

// Execute the complete registered command tree with concrete adapters and an
// in-process TLS Bitbucket endpoint. A new factory prevents command flag and
// BaseRepo state leaking between invocations.
func matrixCLI(t *testing.T, p matrixProvider, srv *httptest.Server, args ...string) (string, error) {
	t.Helper()
	f, out, _ := factorytest.New(t, factorytest.Opts{BackendType: p.name})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(string) string { return srv.URL + p.apiRoot }
	cmd := root.NewCmdRoot(f)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func matrixPR(p matrixProvider, id int, title string) string {
	if p.name == "cloud" {
		return fmt.Sprintf(`{"id":%d,"title":%q,"description":"Review body","state":"OPEN","source":{"branch":{"name":"feature/qa"}},"destination":{"branch":{"name":"main"}},"author":{"account_id":"author-1","display_name":"Author"}}`, id, title)
	}
	return fmt.Sprintf(`{"id":%d,"title":%q,"description":"Review body","state":"OPEN","version":3,"fromRef":{"id":"refs/heads/feature/qa","displayId":"feature/qa"},"toRef":{"id":"refs/heads/main","displayId":"main"},"author":{"user":{"slug":"author-1","displayName":"Author"}},"reviewers":[{"user":{"slug":"qa-reviewer"}}]}`, id, title)
}

func matrixErrorBody(p matrixProvider, message string) string {
	if p.name == "cloud" {
		return fmt.Sprintf(`{"type":"error","error":{"message":%q}}`, message)
	}
	return fmt.Sprintf(`{"errors":[{"message":%q}]}`, message)
}

func TestPublicMatrixPRList(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name+"/success", func(t *testing.T) {
			var requests []string
			var srv *httptest.Server
			srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.Header().Set("Content-Type", "application/json")
				if p.name == "cloud" {
					if r.URL.Query().Get("page") == "2" {
						fmt.Fprintf(w, `{"values":[%s,%s]}`, matrixPR(p, 13, "Third"), matrixPR(p, 14, "Fourth"))
						return
					}
					fmt.Fprintf(w, `{"values":[%s,%s],"next":%q}`, matrixPR(p, 11, "First"), matrixPR(p, 12, "Second"), srv.URL+p.apiRoot+p.prPath+"?state=OPEN&pagelen=3&page=2")
					return
				}
				if r.URL.Query().Get("start") == "2" {
					fmt.Fprintf(w, `{"values":[%s,%s],"isLastPage":true}`, matrixPR(p, 13, "Third"), matrixPR(p, 14, "Fourth"))
					return
				}
				fmt.Fprintf(w, `{"values":[%s,%s],"isLastPage":false,"nextPageStart":2}`, matrixPR(p, 11, "First"), matrixPR(p, 12, "Second"))
			}))
			t.Cleanup(srv.Close)

			out, err := matrixCLI(t, p, srv, "pr", "list", p.host+"/TEAM/service", "--limit", "3", "--json", "--jq", ".[] | .id")
			require.NoError(t, err)
			require.Equal(t, "11\n12\n13\n", out, "the fourth API item must be excluded by --limit")
			require.Len(t, requests, 2)
			require.Contains(t, requests[0], "GET "+p.apiRoot+p.prPath+"?")
			require.Contains(t, requests[0], "state=OPEN")
			if p.name == "cloud" {
				require.Contains(t, requests[0], "pagelen=3")
				require.Contains(t, requests[1], "page=2")
			} else {
				require.Contains(t, requests[0], "limit=3")
				require.Contains(t, requests[1], "start=2")
			}
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr list", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect", "pagination"}, Flags: []string{"--limit", "--json", "--jq"}})
		})

		t.Run(p.name+"/error", func(t *testing.T) {
			var requests []string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, matrixErrorBody(p, "expired token"))
			}))
			t.Cleanup(srv.Close)
			out, err := matrixCLI(t, p, srv, "pr", "list", "--repo", p.host+"/TEAM/service", "--limit", "3")
			require.Error(t, err)
			require.Empty(t, out)
			require.Len(t, requests, 1)
			require.Contains(t, requests[0], "GET "+p.apiRoot+p.prPath+"?")
			var de *backend.DomainError
			require.True(t, errors.As(err, &de))
			require.Equal(t, backend.CodeAuthInvalidToken, de.Code)
			f, _, errOut := factorytest.New(t, factorytest.Opts{})
			errfmt.Render(f.IOStreams, err)
			require.Contains(t, errOut.String(), "Authentication failed")
			require.Contains(t, errOut.String(), "Hint:")
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr list", Backend: p.name, Outcome: "error", Assertions: []string{"input", "effect", "error"}, Flags: []string{"--repo", "--limit"}})
		})
	}
}

func TestPublicMatrixPRCreateAndView(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name+"/success", func(t *testing.T) {
			var created bool
			var posted map[string]any
			var requests []string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == p.apiRoot+p.prPath:
					body, _ := io.ReadAll(r.Body)
					_ = json.Unmarshal(body, &posted)
					created = true
					w.WriteHeader(http.StatusCreated)
					fmt.Fprint(w, matrixPR(p, 47, "Ship QA"))
				case r.Method == http.MethodGet && r.URL.Path == p.apiRoot+p.prPath+"/47" && created:
					fmt.Fprint(w, matrixPR(p, 47, "Ship QA"))
				default:
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
				}
			}))
			t.Cleanup(srv.Close)
			repo := p.host + "/TEAM/service"
			out, err := matrixCLI(t, p, srv, "pr", "create", "--repo", repo, "--title", "Ship QA", "--body", "Review body", "--base", "main", "--head", "feature/qa", "--no-default-reviewers", "--json")
			require.NoError(t, err)
			require.True(t, created)
			require.JSONEq(t, `{"id":47,"title":"Ship QA"}`, selectMatrixFields(t, out, "id", "title"))
			require.Equal(t, "Ship QA", posted["title"])
			require.Equal(t, "Review body", posted["description"])
			if p.name == "cloud" {
				require.Equal(t, "feature/qa", nestedMatrixString(t, posted, "source", "branch", "name"))
				require.Equal(t, "main", nestedMatrixString(t, posted, "destination", "branch", "name"))
			} else {
				require.Equal(t, "refs/heads/feature/qa", nestedMatrixString(t, posted, "fromRef", "id"))
				require.Equal(t, "refs/heads/main", nestedMatrixString(t, posted, "toRef", "id"))
			}
			require.Equal(t, "POST "+p.apiRoot+p.prPath, requests[0])
			view, err := matrixCLI(t, p, srv, "pr", "view", "47", "--repo", repo)
			require.NoError(t, err)
			require.Contains(t, view, "#47 Ship QA")
			require.Contains(t, view, "Review body")
			require.Contains(t, view, "From:   feature/qa")
			require.Len(t, requests, 2)
			require.Equal(t, "GET "+p.apiRoot+p.prPath+"/47", requests[1])
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr create", Backend: p.name, Outcome: "success", Assertions: []string{"output", "effect", "input"}, Flags: []string{"--repo", "--title", "--body", "--base", "--head", "--json", "--no-default-reviewers"}})
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr view", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--repo"}})
		})

		t.Run(p.name+"/create-error", func(t *testing.T) {
			var request string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				request = r.Method + " " + r.URL.RequestURI()
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, matrixErrorBody(p, "pull request already exists"))
			}))
			t.Cleanup(srv.Close)
			out, err := matrixCLI(t, p, srv, "pr", "create", "--repo", p.host+"/TEAM/service", "--title", "Ship QA", "--base", "main", "--head", "feature/qa", "--no-default-reviewers")
			require.Error(t, err)
			require.Empty(t, out)
			require.Equal(t, "POST "+p.apiRoot+p.prPath, request)
			checkMatrixError(t, err, backend.CodePRCreateDuplicateBranch, "already exists", "Hint:")
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr create", Backend: p.name, Outcome: "error", Assertions: []string{"input", "error"}, Flags: []string{"--repo", "--title", "--base", "--head", "--no-default-reviewers"}})
		})

		t.Run(p.name+"/view-error", func(t *testing.T) {
			var requests []string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, matrixErrorBody(p, "pull request missing"))
			}))
			t.Cleanup(srv.Close)
			out, err := matrixCLI(t, p, srv, "pr", "view", "47", "--repo", p.host+"/TEAM/service")
			require.Error(t, err)
			require.Empty(t, out)
			require.Equal(t, []string{"GET " + p.apiRoot + p.prPath + "/47"}, requests)
			checkMatrixError(t, err, backend.CodePRNotFound, "#47", "Hint:")
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr view", Backend: p.name, Outcome: "error", Assertions: []string{"input", "effect", "error"}, Flags: []string{"--repo"}})
		})
	}
}

func checkMatrixError(t *testing.T, err error, code backend.ErrorCode, title, hint string) {
	t.Helper()
	var de *backend.DomainError
	require.True(t, errors.As(err, &de))
	require.Equal(t, code, de.Code)
	f, _, errOut := factorytest.New(t, factorytest.Opts{})
	errfmt.Render(f.IOStreams, err)
	require.Contains(t, errOut.String(), title)
	require.Contains(t, errOut.String(), hint)
}

func nestedMatrixString(t *testing.T, value map[string]any, keys ...string) string {
	t.Helper()
	var current any = value
	for _, key := range keys {
		object, ok := current.(map[string]any)
		require.True(t, ok)
		current = object[key]
	}
	text, ok := current.(string)
	require.True(t, ok)
	return text
}

func selectMatrixFields(t *testing.T, raw string, keys ...string) string {
	t.Helper()
	var value map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(raw)), &value))
	selected := make(map[string]any, len(keys))
	for _, key := range keys {
		selected[key] = value[key]
	}
	b, err := json.Marshal(selected)
	require.NoError(t, err)
	return string(b)
}

func matrixRepo(p matrixProvider, slug string) string {
	if p.name == "cloud" {
		return fmt.Sprintf(`{"full_name":"TEAM/%s","slug":%q,"name":%q,"scm":"git","description":"QA repo","links":{"html":{"href":"https://bitbucket.org/TEAM/%s"}}}`, slug, slug, slug, slug)
	}
	return fmt.Sprintf(`{"id":9,"slug":%q,"name":%q,"scmId":"git","project":{"key":"TEAM"},"links":{"self":[{"href":"https://git.example.test/projects/TEAM/repos/%s"}]}}`, slug, slug, slug)
}

func matrixBranch(p matrixProvider, name string) string {
	if p.name == "cloud" {
		return fmt.Sprintf(`{"name":%q,"target":{"hash":"abcdef1234567890"}}`, name)
	}
	return fmt.Sprintf(`{"id":%q,"displayId":%q,"latestCommit":"abcdef1234567890","isDefault":false}`, "refs/heads/"+name, name)
}

func matrixListPage(p matrixProvider, items []string, next string) string {
	if p.name == "cloud" {
		return fmt.Sprintf(`{"values":[%s],"next":%q}`, strings.Join(items, ","), next)
	}
	if next == "" {
		return fmt.Sprintf(`{"values":[%s],"isLastPage":true}`, strings.Join(items, ","))
	}
	return fmt.Sprintf(`{"values":[%s],"isLastPage":false,"nextPageStart":1}`, strings.Join(items, ","))
}

func TestPublicMatrixRepoViewAndLists(t *testing.T) {
	for _, p := range matrixProviders {
		for _, operation := range []string{"repo view", "repo list", "branch list"} {
			t.Run(p.name+"/"+strings.ReplaceAll(operation, " ", "-")+"/success", func(t *testing.T) {
				var path string
				var args []string
				var first, second string
				var expected []string
				var flags []string
				switch operation {
				case "repo view":
					path = p.apiRoot + matrixRepoPath(p)
					args = []string{"repo", "view", "--repo", p.host + "/TEAM/service", "--json", "--jq", ".slug"}
					first = matrixRepo(p, "service")
					expected = []string{"service"}
					flags = []string{"--repo", "--json", "--jq"}
				case "repo list":
					if p.name == "cloud" {
						path = p.apiRoot + "/repositories/TEAM"
						args = []string{"repo", "list", "TEAM", "--hostname", p.host, "--limit", "2", "--json"}
					} else {
						path = p.apiRoot + "/repos"
						args = []string{"repo", "list", "--hostname", p.host, "--limit", "2", "--json"}
					}
					first, second = matrixRepo(p, "one"), matrixRepo(p, "two")
					expected = []string{"\"slug\":\"one\"", "\"slug\":\"two\""}
					flags = []string{"--hostname", "--limit", "--json"}
				case "branch list":
					path = p.apiRoot + matrixBranchPath(p)
					args = []string{"branch", "list", "--repo", p.host + "/TEAM/service", "--limit", "2", "--json"}
					first, second = matrixBranch(p, "main"), matrixBranch(p, "feature/qa")
					expected = []string{"\"name\":\"main\"", "\"name\":\"feature/qa\""}
					flags = []string{"--repo", "--limit", "--json"}
				}
				var requests []string
				var srv *httptest.Server
				srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests = append(requests, r.Method+" "+r.URL.RequestURI())
					w.Header().Set("Content-Type", "application/json")
					if second == "" {
						fmt.Fprint(w, first)
						return
					}
					if r.URL.Query().Get("page") == "2" || r.URL.Query().Get("start") == "1" {
						fmt.Fprint(w, matrixListPage(p, []string{second, matrixListExtra(p, operation)}, ""))
						return
					}
					next := srv.URL + path + "?page=2"
					fmt.Fprint(w, matrixListPage(p, []string{first}, next))
				}))
				t.Cleanup(srv.Close)
				out, err := matrixCLI(t, p, srv, args...)
				require.NoError(t, err)
				for _, literal := range expected {
					require.Contains(t, out, literal)
				}
				require.Len(t, requests, 1+boolMatrixInt(second != ""))
				require.Contains(t, requests[0], "GET "+path)
				assertions := []string{"output", "input", "effect"}
				if second != "" {
					require.NotContains(t, out, "third")
					if p.name == "cloud" {
						require.Contains(t, requests[0], "pagelen=2")
						require.Contains(t, requests[1], "page=2")
					} else {
						require.Contains(t, requests[0], "limit=2")
						require.Contains(t, requests[1], "start=1")
					}
					assertions = append(assertions, "pagination")
				}
				qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle " + operation, Backend: p.name, Outcome: "success", Assertions: assertions, Flags: flags})
			})

			t.Run(p.name+"/"+strings.ReplaceAll(operation, " ", "-")+"/error", func(t *testing.T) {
				var path string
				var args []string
				var code backend.ErrorCode
				var title string
				switch operation {
				case "repo view":
					path = p.apiRoot + matrixRepoPath(p)
					args = []string{"repo", "view", "--repo", p.host + "/TEAM/service"}
					code, title = backend.CodeRepoNotFound, "Repository TEAM/service"
				case "repo list":
					if p.name == "cloud" {
						path = p.apiRoot + "/repositories/TEAM"
						args = []string{"repo", "list", "TEAM", "--hostname", p.host, "--limit", "2"}
					} else {
						path = p.apiRoot + "/repos"
						args = []string{"repo", "list", "--hostname", p.host, "--limit", "2"}
					}
					code, title = backend.CodeAuthInvalidToken, "Authentication failed"
				case "branch list":
					path = p.apiRoot + matrixBranchPath(p)
					args = []string{"branch", "list", "--repo", p.host + "/TEAM/service", "--limit", "2"}
					code, title = backend.CodeAuthInvalidToken, "Authentication failed"
				}
				var request string
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					request = r.Method + " " + r.URL.RequestURI()
					if operation == "repo view" {
						w.WriteHeader(http.StatusNotFound)
					} else {
						w.WriteHeader(http.StatusUnauthorized)
					}
					fmt.Fprint(w, matrixErrorBody(p, "missing or expired"))
				}))
				t.Cleanup(srv.Close)
				out, err := matrixCLI(t, p, srv, args...)
				require.Error(t, err)
				require.Empty(t, out)
				require.Contains(t, request, "GET "+path)
				checkMatrixError(t, err, code, title, "Hint:")
				flags := []string{"--repo"}
				switch operation {
				case "repo list":
					flags = []string{"--hostname", "--limit"}
				case "branch list":
					flags = []string{"--repo", "--limit"}
				}
				qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle " + operation, Backend: p.name, Outcome: "error", Assertions: []string{"input", "effect", "error"}, Flags: flags})
			})
		}
	}
}

func matrixRepoPath(p matrixProvider) string {
	if p.name == "cloud" {
		return "/repositories/TEAM/service"
	}
	return "/projects/TEAM/repos/service"
}

func matrixBranchPath(p matrixProvider) string {
	if p.name == "cloud" {
		return "/repositories/TEAM/service/refs/branches"
	}
	return "/projects/TEAM/repos/service/branches"
}

func matrixListExtra(p matrixProvider, operation string) string {
	if operation == "repo list" {
		return matrixRepo(p, "third")
	}
	return matrixBranch(p, "third")
}

func boolMatrixInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func TestPublicMatrixBranchWrites(t *testing.T) {
	const hash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	for _, p := range matrixProviders {
		t.Run(p.name+"/create-and-delete", func(t *testing.T) {
			var exists bool
			var createBody map[string]any
			var deleteBody map[string]any
			var requests []string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.Header().Set("Content-Type", "application/json")
				branchPath := p.apiRoot + matrixBranchPath(p)
				switch {
				case r.Method == http.MethodPost && r.URL.Path == branchPath:
					body, _ := io.ReadAll(r.Body)
					_ = json.Unmarshal(body, &createBody)
					exists = true
					w.WriteHeader(http.StatusCreated)
					fmt.Fprint(w, matrixBranch(p, "feature/qa"))
				case r.Method == http.MethodGet && r.URL.Path == branchPath:
					if exists {
						fmt.Fprint(w, matrixListPage(p, []string{matrixBranch(p, "feature/qa")}, ""))
					} else {
						fmt.Fprint(w, matrixListPage(p, nil, ""))
					}
				case r.Method == http.MethodDelete && (r.URL.Path == branchPath || strings.HasPrefix(r.URL.Path, branchPath+"/")):
					body, _ := io.ReadAll(r.Body)
					if len(body) > 0 {
						_ = json.Unmarshal(body, &deleteBody)
					}
					exists = false
					w.WriteHeader(http.StatusNoContent)
				default:
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
				}
			}))
			t.Cleanup(srv.Close)
			repo := p.host + "/TEAM/service"
			created, err := matrixCLI(t, p, srv, "branch", "create", "--repo", repo, "feature/qa", "--start-at", hash)
			require.NoError(t, err)
			require.Equal(t, "Created branch feature/qa\n", created)
			require.Equal(t, "POST "+p.apiRoot+matrixBranchPath(p), requests[0])
			require.Equal(t, "feature/qa", createBody["name"])
			if p.name == "cloud" {
				require.Equal(t, hash, nestedMatrixString(t, createBody, "target", "hash"))
			} else {
				require.Equal(t, hash, createBody["startPoint"])
			}
			listed, err := matrixCLI(t, p, srv, "branch", "list", "--repo", repo, "--json")
			require.NoError(t, err)
			require.Contains(t, listed, `"name":"feature/qa"`)
			require.Contains(t, requests[1], "GET "+p.apiRoot+matrixBranchPath(p))
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle branch create", Backend: p.name, Outcome: "success", Assertions: []string{"output", "effect", "input"}, Flags: []string{"--repo", "--start-at"}})

			deleted, err := matrixCLI(t, p, srv, "branch", "delete", "--repo", repo, "feature/qa")
			require.NoError(t, err)
			require.Equal(t, "Deleted branch feature/qa\n", deleted)
			if p.name == "cloud" {
				require.Equal(t, "DELETE "+p.apiRoot+matrixBranchPath(p)+"/feature%2Fqa", requests[2])
			} else {
				require.Equal(t, "DELETE "+p.apiRoot+matrixBranchPath(p), requests[2])
				require.Equal(t, "feature/qa", deleteBody["name"])
			}
			listed, err = matrixCLI(t, p, srv, "branch", "list", "--repo", repo, "--json")
			require.NoError(t, err)
			require.NotContains(t, listed, "feature/qa")
			require.Contains(t, requests[3], "GET "+p.apiRoot+matrixBranchPath(p))
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle branch delete", Backend: p.name, Outcome: "success", Assertions: []string{"output", "effect", "input"}, Flags: []string{"--repo"}})
		})

		for _, operation := range []string{"create", "delete"} {
			t.Run(p.name+"/"+operation+"-error", func(t *testing.T) {
				var request string
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					request = r.Method + " " + r.URL.RequestURI()
					if operation == "create" {
						w.WriteHeader(http.StatusUnauthorized)
					} else {
						w.WriteHeader(http.StatusForbidden)
					}
					fmt.Fprint(w, matrixErrorBody(p, "branch write denied"))
				}))
				t.Cleanup(srv.Close)
				args := []string{"branch", operation, "--repo", p.host + "/TEAM/service", "feature/qa"}
				flags := []string{"--repo"}
				code, title := backend.CodeBranchProtected, "Branch is protected"
				method := "DELETE "
				if operation == "create" {
					args = append(args, "--start-at", hash)
					flags = append(flags, "--start-at")
					code, title, method = backend.CodeAuthInvalidToken, "Authentication failed", "POST "
				}
				out, err := matrixCLI(t, p, srv, args...)
				require.Error(t, err)
				require.Empty(t, out)
				require.Contains(t, request, method+p.apiRoot+matrixBranchPath(p))
				checkMatrixError(t, err, code, title, "Hint:")
				qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle branch " + operation, Backend: p.name, Outcome: "error", Assertions: []string{"input", "error"}, Flags: flags})
			})
		}
	}
}

func TestPublicMatrixCommitFiles(t *testing.T) {
	for _, p := range matrixProviders {
		path := p.apiRoot + "/projects/TEAM/repos/service/commits/abc123/changes"
		body := `{"values":[{"type":"ADD","path":{"toString":"src/new.go"}}],"isLastPage":true}`
		if p.name == "cloud" {
			path = p.apiRoot + "/repositories/TEAM/service/diffstat/abc123~1..abc123"
			body = `{"values":[{"status":"added","lines_added":5,"lines_removed":0,"new":{"path":"src/new.go"}}]}`
		}
		t.Run(p.name+"/success", func(t *testing.T) {
			var request string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				request = r.Method + " " + r.URL.RequestURI()
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, body)
			}))
			t.Cleanup(srv.Close)
			out, err := matrixCLI(t, p, srv, "commit", "files", "abc123", p.host+"/TEAM/service", "--json")
			require.NoError(t, err)
			require.Equal(t, "GET "+path, request)
			require.Contains(t, out, `"path":"src/new.go"`)
			require.Contains(t, out, `"status":"added"`)
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle commit files", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--json"}})
		})
		t.Run(p.name+"/error", func(t *testing.T) {
			var request string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				request = r.Method + " " + r.URL.RequestURI()
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, matrixErrorBody(p, "expired token"))
			}))
			t.Cleanup(srv.Close)
			out, err := matrixCLI(t, p, srv, "commit", "files", "abc123", p.host+"/TEAM/service")
			require.Error(t, err)
			require.Empty(t, out)
			require.Equal(t, "GET "+path, request)
			checkMatrixError(t, err, backend.CodeAuthInvalidToken, "Authentication failed", "Hint:")
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle commit files", Backend: p.name, Outcome: "error", Assertions: []string{"input", "effect", "error"}})
		})
	}
}

func TestPublicMatrixPREditAndMerge(t *testing.T) {
	for _, p := range matrixProviders {
		for _, operation := range []string{"edit", "merge"} {
			t.Run(p.name+"/"+operation+"/success", func(t *testing.T) {
				var title = "Original title"
				var state = "OPEN"
				var body map[string]any
				var requests []string
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests = append(requests, r.Method+" "+r.URL.RequestURI())
					w.Header().Set("Content-Type", "application/json")
					base := p.apiRoot + p.prPath + "/47"
					switch {
					case r.Method == http.MethodGet && r.URL.Path == base:
						fmt.Fprint(w, matrixPRState(p, title, state))
					case operation == "edit" && r.Method == http.MethodPut && r.URL.Path == base:
						data, _ := io.ReadAll(r.Body)
						_ = json.Unmarshal(data, &body)
						if p.name == "server" && !matrixFullServerPR(body) {
							w.WriteHeader(http.StatusBadRequest)
							fmt.Fprint(w, matrixErrorBody(p, "full PR body with reviewers and branch refs required"))
							return
						}
						title = "Updated title"
						fmt.Fprint(w, matrixPRState(p, title, state))
					case operation == "merge" && r.Method == http.MethodPost && r.URL.Path == base+"/merge":
						data, _ := io.ReadAll(r.Body)
						_ = json.Unmarshal(data, &body)
						state = "MERGED"
						fmt.Fprint(w, matrixPRState(p, title, state))
					default:
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
					}
				}))
				t.Cleanup(srv.Close)
				repo := p.host + "/TEAM/service"
				args := []string{"pr", operation, "47", "--repo", repo}
				flags := []string{"--repo"}
				if operation == "edit" {
					args = append(args, "--title", "Updated title", "--body", "Review body")
					flags = append(flags, "--title", "--body")
				} else {
					args = append(args, "--squash")
					flags = append(flags, "--squash")
				}
				out, err := matrixCLI(t, p, srv, args...)
				require.NoError(t, err)
				require.Equal(t, "GET "+p.apiRoot+p.prPath+"/47", requests[0])
				var writeMethod, writePath string
				if operation == "edit" {
					require.Contains(t, out, "Updated pull request #47")
					writeMethod, writePath = "PUT", p.apiRoot+p.prPath+"/47"
					require.Equal(t, "Updated title", body["title"])
					require.Equal(t, "Review body", body["description"])
				} else {
					require.Equal(t, "Merged pull request #47\n", out)
					writeMethod, writePath = "POST", p.apiRoot+p.prPath+"/47/merge"
					if p.name == "cloud" {
						require.Equal(t, "squash", body["merge_strategy"])
					} else {
						require.Equal(t, "squash", body["strategy"])
					}
				}
				if p.name == "server" {
					require.Equal(t, float64(3), body["version"])
				}
				require.Equal(t, writeMethod+" "+writePath, requests[len(requests)-1])
				view, err := matrixCLI(t, p, srv, "pr", "view", "47", "--repo", repo)
				require.NoError(t, err)
				if operation == "edit" {
					require.Contains(t, view, "#47 Updated title")
				} else {
					require.Contains(t, view, "State:  MERGED")
				}
				require.Equal(t, "GET "+p.apiRoot+p.prPath+"/47", requests[len(requests)-1])
				qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr " + operation, Backend: p.name, Outcome: "success", Assertions: []string{"output", "effect", "input"}, Flags: flags})
			})

			t.Run(p.name+"/"+operation+"/error", func(t *testing.T) {
				var requests []string
				srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests = append(requests, r.Method+" "+r.URL.RequestURI())
					if r.Method == http.MethodGet {
						fmt.Fprint(w, matrixPR(p, 47, "Original title"))
						return
					}
					if operation == "edit" {
						w.WriteHeader(http.StatusUnauthorized)
						fmt.Fprint(w, matrixErrorBody(p, "expired token"))
					} else {
						w.WriteHeader(http.StatusConflict)
						fmt.Fprint(w, matrixErrorBody(p, "source branch is behind"))
					}
				}))
				t.Cleanup(srv.Close)
				args := []string{"pr", operation, "47", "--repo", p.host + "/TEAM/service"}
				flags := []string{"--repo"}
				code, title := backend.CodePRMergeBehind, "behind its target branch"
				method, path := "POST", p.apiRoot+p.prPath+"/47/merge"
				if operation == "edit" {
					args = append(args, "--title", "Updated title")
					flags = append(flags, "--title")
					code, title = backend.CodeAuthInvalidToken, "Authentication failed"
					method, path = "PUT", p.apiRoot+p.prPath+"/47"
				} else {
					args = append(args, "--squash")
					flags = append(flags, "--squash")
				}
				out, err := matrixCLI(t, p, srv, args...)
				require.Error(t, err)
				require.Empty(t, out)
				require.Equal(t, "GET "+p.apiRoot+p.prPath+"/47", requests[0])
				require.Equal(t, method+" "+path, requests[len(requests)-1])
				checkMatrixError(t, err, code, title, "Hint:")
				qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr " + operation, Backend: p.name, Outcome: "error", Assertions: []string{"input", "error"}, Flags: flags})
			})
		}
	}
}

func matrixPRState(p matrixProvider, title, state string) string {
	return strings.Replace(matrixPR(p, 47, title), `"state":"OPEN"`, `"state":"`+state+`"`, 1)
}

func matrixFullServerPR(body map[string]any) bool {
	if body["version"] != float64(3) || body["description"] != "Review body" {
		return false
	}
	from, ok := body["fromRef"].(map[string]any)
	if !ok || from["id"] != "refs/heads/feature/qa" {
		return false
	}
	to, ok := body["toRef"].(map[string]any)
	if !ok || to["id"] != "refs/heads/main" {
		return false
	}
	reviewers, ok := body["reviewers"].([]any)
	if !ok || len(reviewers) != 1 {
		return false
	}
	reviewer, ok := reviewers[0].(map[string]any)
	if !ok {
		return false
	}
	user, ok := reviewer["user"].(map[string]any)
	return ok && user["slug"] == "qa-reviewer"
}
