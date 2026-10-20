package pr_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudQueuePR = `{"id":42,"state":"OPEN","queued":false,"destination":{"branch":{"name":"main"}},"source":{"branch":{"name":"feature/x"},"commit":{"hash":"abc"}},"links":{"html":{"href":"https://bitbucket.org/myworkspace/myrepo/pull-requests/42"}}}`

func newCloudMergeQueueCommand(t *testing.T, srv *httptest.Server) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{Stdout: "git@bitbucket.org:myworkspace/myrepo.git\n"})
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "bitbucket.org:\n  oauth_token: tok\n  git_protocol: ssh\n"})
	f.HTTPClient = factorytest.StubHTTPClient(srv.Client())
	f.BaseURL = func(string) string { return srv.URL }
	f.GitRunner = func() run.Runner { return runner }
	return pr.NewCmdPR(f), out
}

func TestPRMerge_CloudAuto_UsesActiveMergeQueueAndProviderStrategy(t *testing.T) {
	for _, tc := range []struct{ name, strategyFlag, want string }{
		{"merge", "", "merge_commit"},
		{"squash", "--squash", "squash"},
		{"rebase", "--rebase", "rebase_merge"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotStrategy string
			var postCount int
			srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/pullrequests/42":
					_, _ = io.WriteString(w, cloudQueuePR)
				case r.Method == http.MethodGet && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/merge-queues":
					assert.Equal(t, "main", r.URL.Query().Get("target_branch"))
					_, _ = io.WriteString(w, `{"values":[{"uuid":"queue-1","state":"ACTIVE"}]}`)
				case r.Method == http.MethodPost && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/pullrequests/42/merge":
					postCount++
					var body map[string]string
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					gotStrategy = body["merge_strategy"]
					w.WriteHeader(http.StatusAccepted)
				case r.Method == http.MethodGet && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/merge-queues/queue-1/queue/42":
					_, _ = io.WriteString(w, `{"status":"QUEUED","pullrequest":{"id":42},"merge_queue":{"uuid":"queue-1"}}`)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(srv.Close)

			cmd, _ := newCloudMergeQueueCommand(t, srv)
			args := []string{"merge", "42", "--auto"}
			if tc.strategyFlag != "" {
				args = append(args, tc.strategyFlag)
			}
			cmd.SetArgs(args)
			require.NoError(t, cmd.Execute())
			assert.Equal(t, tc.want, gotStrategy)
			assert.Equal(t, 1, postCount)
			flags := []string{"--repo", "--auto"}
			if tc.strategyFlag != "" {
				flags = append(flags, tc.strategyFlag)
			}
			qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr merge", Backend: "cloud", Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: flags})
		})
	}
}

func TestPRMerge_CloudAuto_WithoutActiveQueueDoesNotSubmitMerge(t *testing.T) {
	postCalled := false
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/pullrequests/42":
			_, _ = io.WriteString(w, cloudQueuePR)
		case r.Method == http.MethodGet && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/merge-queues":
			_, _ = io.WriteString(w, `{"values":[]}`)
		case r.Method == http.MethodPost:
			postCalled = true
			t.Errorf("merge must not be submitted without an active queue")
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newCloudMergeQueueCommand(t, srv)
	cmd.SetArgs([]string{"merge", "42", "--auto"})
	err := cmd.Execute()
	require.Error(t, err)
	var domainErr *backend.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, backend.CodePRMergeQueueRequired, domainErr.Code)
	assert.False(t, postCalled)
	qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr merge", Backend: "cloud", Outcome: "error", Assertions: []string{"error", "input", "effect"}, Flags: []string{"--repo", "--auto"}})
}

func TestPRMerge_CloudAutoOff_CancelsFromDrainingQueueAndConfirmsOpenPR(t *testing.T) {
	removed := false
	var deleteCount int
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/pullrequests/42":
			prJSON := cloudQueuePR
			if removed {
				prJSON = `{"id":42,"state":"OPEN","queued":false,"destination":{"branch":{"name":"main"}},"source":{"branch":{"name":"feature/x"},"commit":{"hash":"abc"}},"links":{"html":{"href":"https://bitbucket.org/myworkspace/myrepo/pull-requests/42"}}}`
			}
			_, _ = io.WriteString(w, prJSON)
		case r.Method == http.MethodGet && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/merge-queues":
			_, _ = io.WriteString(w, `{"values":[{"uuid":"queue-draining","state":"DRAINING"}]}`)
		case r.Method == http.MethodGet && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/merge-queues/queue-draining/queue/42":
			if removed {
				http.NotFound(w, r)
				return
			}
			_, _ = io.WriteString(w, `{"status":"QUEUED","pullrequest":{"id":42},"merge_queue":{"uuid":"queue-draining"}}`)
		case r.Method == http.MethodDelete && r.URL.Path == "/repositories/MYWORKSPACE/myrepo/merge-queues/queue-draining/queue/42":
			deleteCount++
			removed = true
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	cmd, _ := newCloudMergeQueueCommand(t, srv)
	cmd.SetArgs([]string{"merge", "42", "--auto-off"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 1, deleteCount)
	assert.True(t, removed)
	qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr merge", Backend: "cloud", Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--repo", "--auto-off"}})
}

func TestPRView_CloudReportsQueueMembershipWithoutUnknownStrategy(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repositories/MYWORKSPACE/myrepo/pullrequests/42" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":42,"title":"Queued PR","state":"OPEN","queued":true,"destination":{"branch":{"name":"main"}},"source":{"branch":{"name":"feature/x"},"commit":{"hash":"abc"}},"links":{"html":{"href":"https://bitbucket.org/myworkspace/myrepo/pull-requests/42"}}}`)
	}))
	t.Cleanup(srv.Close)

	cmd, out := newCloudMergeQueueCommand(t, srv)
	cmd.SetArgs([]string{"view", "42"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Auto-merge: enabled")
	assert.NotContains(t, out.String(), "enabled ()")
	qamatrix.Record(t, qamatrix.Scenario{Operation: "cli:bitbottle pr view", Backend: "cloud", Outcome: "success", Assertions: []string{"output", "input", "effect"}, Flags: []string{"--repo"}})
}
