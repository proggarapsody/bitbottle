package contract_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
)

// Invalid merge requests must fail locally, before contacting either provider.
func TestPublicCriticalPRMergeRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name      string
		args      []string
		wantError string
	}{
		{"merge_and_squash", []string{"47", "--merge", "--squash"}, "choose only one of --merge, --squash, or --rebase"},
		{"merge_and_rebase", []string{"47", "--merge", "--rebase"}, "choose only one of --merge, --squash, or --rebase"},
		{"squash_and_rebase", []string{"47", "--squash", "--rebase"}, "choose only one of --merge, --squash, or --rebase"},
		{"enable_and_cancel_auto_merge", []string{"47", "--auto", "--auto-off"}, "cannot use --auto and --auto-off together"},
		{"missing_id", []string{"--merge"}, "accepts 1 arg(s), received 0"},
		{"nonnumeric_id", []string{"not-a-pr", "--merge"}, `invalid PR ID "not-a-pr": must be a positive integer`},
	}
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					var requests atomic.Int32
					srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests.Add(1)
						w.WriteHeader(http.StatusInternalServerError)
					}))
					t.Cleanup(srv.Close)
					args := append([]string{"pr", "merge", "--repo", p.host + "/TEAM/service"}, tc.args...)

					out, err := matrixCLI(t, p, srv, args...)

					assert.EqualError(t, err, tc.wantError)
					assert.Zero(t, requests.Load(), "invalid input must not reach the provider")
					assert.Empty(t, out, "invalid input must not print success")
				})
			}
		})
	}
}

// --delete-branch applies only after a successful merge (README.md's public
// merge example); a denied or conflicted merge must leave both resources intact.
func TestPublicCriticalPRMergeFailurePreservesSourceBranch(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		message string
		code    backend.ErrorCode
	}{
		// The generic 403 classifier promises ErrPermission, without a dotted code.
		{"forbidden", http.StatusForbidden, "write permission required", ""},
		{"conflict", http.StatusConflict, "merge conflict", backend.CodePRMergeConflict},
	}
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					prPath := p.apiRoot + p.prPath + "/47"
					branchPath := p.apiRoot + matrixBranchPath(p)
					branchExists := true
					var deletes, merges int
					srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						switch {
						case r.Method == http.MethodGet && r.URL.Path == prPath:
							fmt.Fprint(w, matrixPR(p, 47, "Unmerged feature"))
						case r.Method == http.MethodPost && r.URL.Path == prPath+"/merge":
							merges++
							w.WriteHeader(tc.status)
							fmt.Fprint(w, matrixErrorBody(p, tc.message))
						case r.Method == http.MethodDelete:
							// Observe any attempted deletion, including an empty branch
							// name after a lost merge error; never hide it behind a 404.
							deletes++
							branchExists = false
							w.WriteHeader(http.StatusNoContent)
						case r.Method == http.MethodGet && r.URL.Path == branchPath:
							var branches []string
							if branchExists {
								branches = []string{matrixBranch(p, "feature/qa")}
							}
							fmt.Fprint(w, matrixListPage(p, branches, ""))
						default:
							w.WriteHeader(http.StatusNotFound)
						}
					}))
					t.Cleanup(srv.Close)
					repo := p.host + "/TEAM/service"

					out, err := matrixCLI(t, p, srv, "pr", "merge", "47", "--repo", repo, "--merge", "--delete-branch")

					assert.Empty(t, out, "failed merge must not print success")
					assert.Zero(t, deletes, "failed merge must not attempt source branch deletion")
					assert.Equal(t, 1, merges, "exercise the provider merge failure")
					var domainErr *backend.DomainError
					if assert.ErrorAs(t, err, &domainErr) {
						if tc.code != "" {
							assert.Equal(t, tc.code, domainErr.Code)
						}
						assert.Equal(t, tc.status, domainErr.HTTPStatus())
						if tc.status == http.StatusForbidden {
							assert.ErrorIs(t, err, backend.ErrPermission)
						}
					}
					view, err := matrixCLI(t, p, srv, "pr", "view", "47", "--repo", repo, "--json")
					require.NoError(t, err)
					var pr struct {
						State string `json:"state"`
					}
					require.NoError(t, json.Unmarshal([]byte(view), &pr))
					assert.Equal(t, "OPEN", pr.State)
					branches, err := matrixCLI(t, p, srv, "branch", "list", "--repo", repo, "--json", "--jq", ".[].name")
					require.NoError(t, err)
					assert.Equal(t, "\"feature/qa\"\n", branches)
					if !t.Failed() {
						qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr merge", Backend: p.name, Outcome: "error", Assertions: []string{"error", "effect", "output"}, Flags: []string{"--merge", "--delete-branch"}})
					}
				})
			}
		})
	}
}
