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

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
)

type matrixAPIClient struct{ server *httptest.Server }

func (c matrixAPIClient) Do(req *http.Request) (*http.Response, error) {
	copyReq := req.Clone(req.Context())
	copyReq.URL.Scheme = "https"
	copyReq.URL.Host = strings.TrimPrefix(c.server.URL, "https://")
	copyReq.Host = copyReq.URL.Host
	return c.server.Client().Do(copyReq)
}

func matrixServerAPI(t *testing.T, p matrixProvider, srv *httptest.Server, endpoint string) (string, error) {
	t.Helper()
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: p.host + ":\n  oauth_token: test-token\n"})
	f.HTTPClient = factorytest.StubHTTPClient(matrixAPIClient{server: srv})
	cmd := root.NewCmdRoot(f)
	cmd.SetArgs([]string{"api", strings.TrimPrefix(p.apiRoot, "/") + endpoint, "--hostname", p.host})
	err := cmd.Execute()
	return out.String(), err
}

func TestPublicMatrixBranchCreateJSON(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			var created bool
			var reads int
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				path := p.apiRoot + matrixBranchPath(p)
				switch {
				case r.Method == http.MethodPost && r.URL.Path == path:
					created = true
					w.WriteHeader(http.StatusCreated)
					fmt.Fprint(w, matrixBranch(p, "feature/qa"))
				case r.Method == http.MethodGet && r.URL.Path == path:
					reads++
					if created {
						fmt.Fprint(w, matrixListPage(p, []string{matrixBranch(p, "feature/qa")}, ""))
					} else {
						fmt.Fprint(w, matrixListPage(p, nil, ""))
					}
				default:
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
				}
			}))
			t.Cleanup(srv.Close)
			repo := p.host + "/TEAM/service"
			out, err := matrixCLI(t, p, srv, "branch", "create", "--repo", repo, "feature/qa", "--start-at", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "--json")
			require.NoError(t, err)
			require.JSONEq(t, `{"name":"feature/qa","default":false,"hash":"abcdef12"}`, out)
			listed, err := matrixCLI(t, p, srv, "branch", "list", "--repo", repo, "--json")
			require.NoError(t, err)
			require.Contains(t, listed, `"name":"feature/qa"`)
			require.Equal(t, 1, reads)
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle branch create", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--json", "--repo", "--start-at"}})
		})
	}
}

func TestPublicMatrixPRStateAndJSON(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			var requests []string
			merged := strings.Replace(matrixPR(p, 47, "Merged feature"), `"state":"OPEN"`, `"state":"MERGED"`, 1)
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet && r.URL.Path == p.apiRoot+p.prPath+"/47" {
					fmt.Fprint(w, merged)
					return
				}
				if r.Method == http.MethodGet && r.URL.Path == p.apiRoot+p.prPath && r.URL.Query().Get("state") == "MERGED" {
					fmt.Fprint(w, matrixListPage(p, []string{merged}, ""))
					return
				}
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
			}))
			t.Cleanup(srv.Close)
			repo := p.host + "/TEAM/service"
			listed, err := matrixCLI(t, p, srv, "pr", "list", "--repo", repo, "--state", "merged", "--limit", "1", "--json")
			require.NoError(t, err)
			require.Contains(t, listed, `"title":"Merged feature"`)
			require.Contains(t, requests[0], "state=MERGED")
			if p.name == "cloud" {
				require.Contains(t, requests[0], "pagelen=1")
			} else {
				require.Contains(t, requests[0], "limit=1")
			}
			require.Len(t, requests, 1)
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr list", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect", "pagination"}, Flags: []string{"--state"}})
			view, err := matrixCLI(t, p, srv, "pr", "view", "47", "--repo", repo, "--json")
			require.NoError(t, err)
			require.JSONEq(t, `{"id":47,"title":"Merged feature","state":"MERGED"}`, selectMatrixFields(t, view, "id", "title", "state"))
			require.Equal(t, "GET "+p.apiRoot+p.prPath+"/47", requests[1])
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr view", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--json"}})
		})
	}
}

func TestPublicMatrixPREditRemoveReviewer(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			reviewers := map[string]bool{"qa-reviewer": true, "other-reviewer": true}
			var writes []string
			var cloudPut map[string]any
			prBody := func() string {
				var pr map[string]any
				require.NoError(t, json.Unmarshal([]byte(matrixPR(p, 47, "Reviewers")), &pr))
				entries := []any{}
				for _, name := range []string{"qa-reviewer", "other-reviewer"} {
					if !reviewers[name] {
						continue
					}
					if p.name == "cloud" {
						entries = append(entries, map[string]any{"account_id": name})
					} else {
						entries = append(entries, map[string]any{"user": map[string]any{"slug": name, "name": name}})
					}
				}
				pr["reviewers"] = entries
				body, err := json.Marshal(pr)
				require.NoError(t, err)
				return string(body)
			}
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				base := p.apiRoot + p.prPath + "/47"
				switch {
				case r.Method == http.MethodGet && r.URL.Path == base:
					fmt.Fprint(w, prBody())
				case p.name == "cloud" && r.Method == http.MethodPut && r.URL.Path == base:
					data, _ := io.ReadAll(r.Body)
					_ = json.Unmarshal(data, &cloudPut)
					writes = append(writes, r.Method+" "+r.URL.RequestURI())
					entries, _ := cloudPut["reviewers"].([]any)
					reviewers = map[string]bool{}
					for _, entry := range entries {
						reviewers[entry.(map[string]any)["account_id"].(string)] = true
					}
					fmt.Fprint(w, prBody())
				case p.name == "server" && r.Method == http.MethodDelete && r.URL.Path == base+"/participants/qa-reviewer":
					writes = append(writes, r.Method+" "+r.URL.RequestURI())
					delete(reviewers, "qa-reviewer")
					w.WriteHeader(http.StatusNoContent)
				case r.Method == http.MethodGet && r.URL.Path == base+"/participants":
					items := []string{}
					for _, name := range []string{"qa-reviewer", "other-reviewer"} {
						if !reviewers[name] {
							continue
						}
						if p.name == "cloud" {
							items = append(items, fmt.Sprintf(`{"user":{"account_id":%q},"role":"REVIEWER"}`, name))
						} else {
							items = append(items, fmt.Sprintf(`{"user":{"slug":%q},"role":"REVIEWER"}`, name))
						}
					}
					fmt.Fprint(w, matrixListPage(p, items, ""))
				default:
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
				}
			}))
			t.Cleanup(srv.Close)
			repo := p.host + "/TEAM/service"
			out, err := matrixCLI(t, p, srv, "pr", "edit", "47", "--repo", repo, "--remove-reviewer", "qa-reviewer")
			require.NoError(t, err)
			require.Contains(t, out, "Removed reviewer(s) from pull request #47")
			require.Len(t, writes, 1)
			if p.name == "cloud" {
				require.Equal(t, "PUT "+p.apiRoot+p.prPath+"/47", writes[0])
				require.Equal(t, "Reviewers", cloudPut["title"])
				require.Equal(t, []any{map[string]any{"account_id": "other-reviewer"}}, cloudPut["reviewers"])
			} else {
				require.Equal(t, "DELETE "+p.apiRoot+p.prPath+"/47/participants/qa-reviewer", writes[0])
			}
			listed, err := matrixCLI(t, p, srv, "pr", "participants", "47", "--repo", repo, "--json")
			require.NoError(t, err)
			require.Contains(t, listed, `"username":"other-reviewer"`)
			require.NotContains(t, listed, "qa-reviewer")
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr edit", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--remove-reviewer"}})
		})
	}
}

func TestPublicMatrixPRMergeCommitAndDeleteBranch(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			state, branchExists := "OPEN", true
			var mergeBody, deleteBody map[string]any
			var requests []string
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests = append(requests, r.Method+" "+r.URL.RequestURI())
				w.Header().Set("Content-Type", "application/json")
				base := p.apiRoot + p.prPath + "/47"
				branchPath := p.apiRoot + matrixBranchPath(p)
				switch {
				case r.Method == http.MethodGet && r.URL.Path == base:
					fmt.Fprint(w, matrixPRState(p, "Merge feature", state))
				case r.Method == http.MethodPost && r.URL.Path == base+"/merge":
					data, _ := io.ReadAll(r.Body)
					_ = json.Unmarshal(data, &mergeBody)
					state = "MERGED"
					fmt.Fprint(w, matrixPRState(p, "Merge feature", state))
				case r.Method == http.MethodDelete && (r.URL.Path == branchPath || strings.HasPrefix(r.URL.Path, branchPath+"/")):
					data, _ := io.ReadAll(r.Body)
					if len(data) > 0 {
						_ = json.Unmarshal(data, &deleteBody)
					}
					branchExists = false
					w.WriteHeader(http.StatusNoContent)
				case r.Method == http.MethodGet && r.URL.Path == branchPath:
					items := []string(nil)
					if branchExists {
						items = []string{matrixBranch(p, "feature/qa")}
					}
					fmt.Fprint(w, matrixListPage(p, items, ""))
				default:
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
				}
			}))
			t.Cleanup(srv.Close)
			repo := p.host + "/TEAM/service"
			out, err := matrixCLI(t, p, srv, "pr", "merge", "47", "--repo", repo, "--merge", "--delete-branch")
			require.NoError(t, err)
			require.Equal(t, "Merged pull request #47\n", out)
			if p.name == "cloud" {
				require.Equal(t, "merge_commit", mergeBody["merge_strategy"])
				require.Contains(t, requests, "DELETE "+p.apiRoot+matrixBranchPath(p)+"/feature%2Fqa")
			} else {
				require.Equal(t, "no-ff", mergeBody["strategyId"])
				require.NotContains(t, mergeBody, "strategy")
				require.Equal(t, float64(3), mergeBody["version"])
				require.Equal(t, "feature/qa", deleteBody["name"])
			}
			view, err := matrixCLI(t, p, srv, "pr", "view", "47", "--repo", repo, "--json")
			require.NoError(t, err)
			require.Contains(t, view, `"state":"MERGED"`)
			branches, err := matrixCLI(t, p, srv, "branch", "list", "--repo", repo, "--json")
			require.NoError(t, err)
			require.NotContains(t, branches, "feature/qa")
			flags := []string{"--merge", "--delete-branch"}
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr merge", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: flags})
		})
	}
}

func TestPublicMatrixPRAutoMergeSquashAndCancel(t *testing.T) {
	p := matrixProviders[1] // Server/DC auto-merge request is documented by its REST contract.
	t.Run(p.name, func(t *testing.T) {
		enabled := false
		var queued map[string]any
		var requests []string
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests = append(requests, r.Method+" "+r.URL.RequestURI())
			w.Header().Set("Content-Type", "application/json")
			base := p.apiRoot + p.prPath + "/47"
			switch {
			case r.Method == http.MethodPost && r.URL.Path == base+"/merge":
				data, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(data, &queued)
				enabled = true
				fmt.Fprint(w, `{}`)
			case r.Method == http.MethodDelete && r.URL.Path == base+"/auto-merge":
				enabled = false
				w.WriteHeader(http.StatusNoContent)
			case r.Method == http.MethodGet && r.URL.Path == base:
				fmt.Fprint(w, matrixPR(p, 47, "Auto feature"))
			case r.Method == http.MethodGet && r.URL.Path == base+"/auto-merge" && enabled:
				fmt.Fprint(w, `{"strategyId":"squash"}`)
			case r.Method == http.MethodGet && r.URL.Path == base+"/auto-merge":
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, matrixErrorBody(p, "auto-merge is not queued"))
			default:
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
			}
		}))
		t.Cleanup(srv.Close)
		repo := p.host + "/TEAM/service"
		out, err := matrixCLI(t, p, srv, "pr", "merge", "47", "--repo", repo, "--auto", "--squash")
		require.NoError(t, err)
		require.Equal(t, "Auto-merge requested for pull request #47 (squash)\n", out)
		require.Equal(t, "squash", queued["strategyId"])
		require.Equal(t, true, queued["autoMerge"])
		require.Equal(t, float64(3), queued["version"])
		require.Equal(t, "GET "+p.apiRoot+p.prPath+"/47", requests[0])
		require.Equal(t, "POST "+p.apiRoot+p.prPath+"/47/merge", requests[1])
		view, err := matrixServerAPI(t, p, srv, p.prPath+"/47/auto-merge")
		require.NoError(t, err)
		require.JSONEq(t, `{"strategyId":"squash"}`, view)
		qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr merge", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--auto", "--squash"}})

		out, err = matrixCLI(t, p, srv, "pr", "merge", "47", "--repo", repo, "--auto-off")
		require.NoError(t, err)
		require.Equal(t, "Cancelled auto-merge for pull request #47\n", out)
		require.Contains(t, requests, "DELETE "+p.apiRoot+p.prPath+"/47/auto-merge")
		view, err = matrixServerAPI(t, p, srv, p.prPath+"/47/auto-merge")
		require.Error(t, err)
		require.Contains(t, view, "auto-merge is not queued")
		qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr merge", Backend: p.name, Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--auto-off"}})
	})
}
