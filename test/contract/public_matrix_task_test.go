package contract_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
)

func TestPublicMatrixTaskCreateCapability(t *testing.T) {
	for _, p := range matrixProviders {
		t.Run(p.name, func(t *testing.T) {
			repo := p.host + "/TEAM/service"
			posted := 0
			deniedWrites := 0
			taskCreated := false
			anchoredCreated := false
			taskPath := p.apiRoot + p.prPath + "/42"
			var postedBody map[string]any
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if p.name == "cloud" {
					switch {
					case r.Method == http.MethodGet && r.URL.Path == taskPath+"/comments":
						fmt.Fprint(w, `{"values":[{"id":11,"content":{"raw":"Existing note"}}]}`)
					case r.Method == http.MethodPost && r.URL.Path == taskPath+"/comments":
						posted++
						w.WriteHeader(http.StatusCreated)
						fmt.Fprint(w, `{"id":13,"content":{"raw":"Fix this"}}`)
					default:
						w.WriteHeader(http.StatusNotFound)
						fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
					}
					return
				}
				switch {
				case r.Method == http.MethodPost && r.URL.Path == taskPath+"/comments":
					posted++
					require.NoError(t, json.NewDecoder(r.Body).Decode(&postedBody))
					if postedBody["text"] == "Denied task" {
						deniedWrites++
						w.WriteHeader(http.StatusForbidden)
						fmt.Fprint(w, matrixErrorBody(p, "write denied"))
						return
					}
					taskCreated = true
					w.WriteHeader(http.StatusCreated)
					if posted == 2 {
						anchoredCreated = true
						fmt.Fprint(w, `{"id":14,"text":"Reply task","severity":"BLOCKER","state":"OPEN","version":1}`)
						return
					}
					fmt.Fprint(w, `{"id":13,"text":"Fix this","severity":"BLOCKER","state":"OPEN","version":1}`)
				case r.Method == http.MethodGet && r.URL.Path == taskPath+"/activities":
					if !taskCreated {
						fmt.Fprint(w, `{"values":[{"action":"COMMENTED","comment":{"id":11,"text":"Existing note"}}],"isLastPage":true}`)
						return
					}
					if anchoredCreated {
						fmt.Fprint(w, `{"values":[{"action":"COMMENTED","comment":{"id":13,"text":"Fix this","severity":"BLOCKER","state":"OPEN","version":1}},{"action":"COMMENTED","comment":{"id":14,"text":"Reply task","severity":"BLOCKER","state":"OPEN","version":1,"parent":{"id":11}}}],"isLastPage":true}`)
						return
					}
					fmt.Fprint(w, `{"values":[{"action":"COMMENTED","comment":{"id":13,"text":"Fix this","severity":"BLOCKER","state":"OPEN","version":1}}],"isLastPage":true}`)
				default:
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, matrixErrorBody(p, "unexpected route"))
				}
			}))
			t.Cleanup(srv.Close)
			if p.name == "cloud" {
				before, err := matrixCLI(t, p, srv, "pr", "comment", "list", "42", "--repo", repo, "--json")
				require.NoError(t, err)
				require.Contains(t, before, "Existing note")
				out, err := matrixCLI(t, p, srv, "pr", "task", "create", "42", "--repo", repo, "--body", "Fix this")
				require.Error(t, err)
				require.Empty(t, out)
				var de *backend.DomainError
				require.True(t, errors.As(err, &de))
				require.Equal(t, backend.CodeHostUnsupported, de.Code)
				require.Equal(t, 0, posted, "unsupported task must not write an ordinary Cloud comment")
				after, err := matrixCLI(t, p, srv, "pr", "comment", "list", "42", "--repo", repo, "--json")
				require.NoError(t, err)
				require.JSONEq(t, before, after)
				qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr task create", Backend: "cloud", Outcome: "unsupported", Assertions: []string{"error", "input", "effect"}, Flags: []string{"--repo", "--body"}})
				return
			}
			ordinary, err := matrixCLI(t, p, srv, "pr", "comment", "list", "42", "--repo", repo, "--json")
			require.NoError(t, err)
			require.Contains(t, ordinary, "Existing note")
			emptyTasks, err := matrixCLI(t, p, srv, "pr", "task", "list", "42", "--repo", repo, "--json")
			require.NoError(t, err)
			require.JSONEq(t, "[]", emptyTasks)
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr task list", Backend: "server", Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--repo", "--json"}})
			out, err := matrixCLI(t, p, srv, "pr", "task", "create", "42", "--repo", repo, "--body", "Fix this")
			require.NoError(t, err)
			require.Contains(t, out, "Created task #13")
			require.Equal(t, 1, posted)
			require.Equal(t, "Fix this", postedBody["text"])
			require.Equal(t, "BLOCKER", postedBody["severity"])
			list, err := matrixCLI(t, p, srv, "pr", "task", "list", "42", "--repo", repo, "--json")
			require.NoError(t, err)
			require.Contains(t, list, "Fix this")
			require.Contains(t, list, "BLOCKER")
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr task create", Backend: "server", Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--repo", "--body"}})
			anchoredOut, err := matrixCLI(t, p, srv, "pr", "task", "create", "42", "--repo", repo, "--body", "Reply task", "--parent", "11")
			require.NoError(t, err)
			require.Contains(t, anchoredOut, "Created task #14")
			require.Equal(t, 2, posted)
			require.Equal(t, "Reply task", postedBody["text"])
			require.Equal(t, "BLOCKER", postedBody["severity"])
			require.Equal(t, float64(11), postedBody["parent"].(map[string]any)["id"])
			list, err = matrixCLI(t, p, srv, "pr", "task", "list", "42", "--repo", repo, "--json")
			require.NoError(t, err)
			require.Contains(t, list, "Reply task")
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr task create", Backend: "server", Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--repo", "--body", "--parent"}})
			beforeDenied := list
			deniedOut, err := matrixCLI(t, p, srv, "pr", "task", "create", "42", "--repo", repo, "--body", "Denied task")
			require.Error(t, err)
			require.Empty(t, deniedOut)
			var de *backend.DomainError
			require.True(t, errors.As(err, &de))
			require.ErrorIs(t, de, backend.ErrPermission)
			require.Equal(t, http.StatusForbidden, de.HTTPStatus())
			require.Equal(t, 3, posted)
			require.Equal(t, 1, deniedWrites)
			require.Equal(t, "Denied task", postedBody["text"])
			list, err = matrixCLI(t, p, srv, "pr", "task", "list", "42", "--repo", repo, "--json")
			require.NoError(t, err)
			require.JSONEq(t, beforeDenied, list)
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr task create", Backend: "server", Outcome: "error", Assertions: []string{"error", "input", "effect"}, Flags: []string{"--repo", "--body"}})
		})
	}
}
