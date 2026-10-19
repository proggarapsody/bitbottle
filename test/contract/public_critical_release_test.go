package contract_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
)

func TestPublicCriticalBranchListStopsAtTotalLimit(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			path := p.apiRoot + matrixBranchPath(p)
			var requests []string
			var srv *httptest.Server
			srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.Header().Set("Content-Type", "application/json")
				if r.Method != http.MethodGet || r.URL.Path != path {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				if p.name == "cloud" {
					switch r.URL.Query().Get("page") {
					case "":
						fmt.Fprintf(w, `{"values":[%s,%s],"next":%q}`, matrixBranch(p, "main"), matrixBranch(p, "feature/first"), srv.URL+path+"?page=2")
					case "2":
						fmt.Fprintf(w, `{"values":[%s,%s],"next":%q}`, matrixBranch(p, "feature/sentinel"), matrixBranch(p, "feature/over-limit"), srv.URL+path+"?page=3")
					default:
						w.WriteHeader(http.StatusInternalServerError)
					}
					return
				}
				switch r.URL.Query().Get("start") {
				case "":
					fmt.Fprintf(w, `{"values":[%s,%s],"isLastPage":false,"nextPageStart":2}`, matrixBranch(p, "main"), matrixBranch(p, "feature/first"))
				case "2":
					fmt.Fprintf(w, `{"values":[%s,%s],"isLastPage":false,"nextPageStart":4}`, matrixBranch(p, "feature/sentinel"), matrixBranch(p, "feature/over-limit"))
				default:
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			t.Cleanup(srv.Close)

			out, err := matrixCLI(t, p, srv, "branch", "list", "--repo", p.host+"/TEAM/service", "--limit", "3", "--json", "--jq", ".[].name")
			require.NoError(t, err)
			require.Equal(t, "\"main\"\n\"feature/first\"\n\"feature/sentinel\"\n", out)
			if p.name == "cloud" {
				require.Equal(t, []string{
					"GET " + path + "?pagelen=3",
					"GET " + path + "?page=2",
				}, requests)
			} else {
				require.Equal(t, []string{
					"GET " + path + "?limit=3",
					"GET " + path + "?limit=3&start=2",
				}, requests)
			}
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle branch list", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect", "pagination"}, Flags: []string{"--repo", "--limit", "--json", "--jq"}})
		})
	}
}

func TestPublicCriticalForbiddenBranchDeletePreservesBranch(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			path := p.apiRoot + matrixBranchPath(p)
			var requests []string
			var deleteBody map[string]any
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodDelete && (r.URL.Path == path || strings.HasPrefix(r.URL.Path, path+"/")):
					if p.name == "server" {
						body, _ := io.ReadAll(r.Body)
						_ = json.Unmarshal(body, &deleteBody)
					}
					w.WriteHeader(http.StatusForbidden)
					fmt.Fprint(w, matrixErrorBody(p, "protected branch"))
				case r.Method == http.MethodGet && r.URL.Path == path:
					fmt.Fprint(w, matrixListPage(p, []string{matrixBranch(p, "release/stable")}, ""))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			t.Cleanup(srv.Close)

			repo := p.host + "/TEAM/service"
			out, err := matrixCLI(t, p, srv, "branch", "delete", "--repo", repo, "release/stable")
			require.Error(t, err)
			require.Empty(t, out)
			checkMatrixError(t, err, backend.CodeBranchProtected, "Branch is protected", "Hint:")
			listed, err := matrixCLI(t, p, srv, "branch", "list", "--repo", repo, "--json", "--jq", ".[].name")
			require.NoError(t, err)
			require.Equal(t, "\"release/stable\"\n", listed)
			if p.name == "cloud" {
				require.Equal(t, "DELETE "+path+"/release%2Fstable", requests[0])
			} else {
				require.Equal(t, "DELETE "+path, requests[0])
				require.Equal(t, "release/stable", deleteBody["name"])
			}
			require.Len(t, requests, 2)
			require.True(t, strings.HasPrefix(requests[1], "GET "+path+"?"), requests[1])
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle branch delete", Backend: p.name, Outcome: "error", Assertions: []string{"error", "input", "effect"}, Flags: []string{"--repo"}})
		})
	}
}

func TestPublicCriticalCommitFilesFromCurrentRepo(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			path := p.apiRoot + "/projects/TEAM/repos/service/commits/abc123/changes"
			body := `{"values":[{"type":"ADD","path":{"toString":"src/new.go"}},{"type":"DELETE","path":{"toString":"src/old.go"}}],"isLastPage":true}`
			if p.name == "cloud" {
				path = p.apiRoot + "/repositories/TEAM/service/diffstat/abc123~1..abc123"
				body = `{"values":[{"status":"added","lines_added":5,"lines_removed":0,"new":{"path":"src/new.go"}},{"status":"removed","lines_added":0,"lines_removed":3,"old":{"path":"src/old.go"}}]}`
			}
			var requests []string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.Header().Set("Content-Type", "application/json")
				if r.Method != http.MethodGet || r.URL.Path != path {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				fmt.Fprint(w, body)
			}))
			t.Cleanup(srv.Close)

			run := func(args ...string) (string, error) {
				f, out, _ := factorytest.New(t, factorytest.Opts{BackendType: p.name})
				f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
				f.BaseURL = func(string) string { return srv.URL + p.apiRoot }
				f.BaseRepo = func() (bbrepo.RepoRef, error) {
					return bbrepo.RepoRef{Host: p.host, Project: "TEAM", Slug: "service"}, nil
				}
				cmd := root.NewCmdRoot(f)
				cmd.SetArgs(args)
				err := cmd.Execute()
				return out.String(), err
			}

			out, err := run("commit", "files", "abc123", "-k", "--json")
			require.NoError(t, err)
			var entries []struct {
				Path   string `json:"path"`
				Status string `json:"status"`
			}
			require.NoError(t, json.Unmarshal([]byte(out), &entries))
			require.Equal(t, 2, len(entries))
			require.Equal(t, "src/new.go", entries[0].Path)
			require.Equal(t, "added", entries[0].Status)
			require.Equal(t, "src/old.go", entries[1].Path)
			require.Equal(t, "deleted", entries[1].Status)

			filtered, err := run("commit", "files", "abc123", "-k", "--json", "--jq", ".[] | .path")
			require.NoError(t, err)
			require.Equal(t, "\"src/new.go\"\n\"src/old.go\"\n", filtered)
			require.Equal(t, []string{"GET " + path, "GET " + path}, requests)
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle commit files", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--json", "--jq"}})
		})
	}
}
