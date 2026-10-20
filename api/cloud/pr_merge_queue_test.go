package cloud_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestEnableAutoMerge_RequiresActiveQueueBeforeAnyWrite(t *testing.T) {
	writes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writes++
			w.WriteHeader(202)
			return
		}
		switch r.URL.Path {
		case "/repositories/ws/repo/pullrequests/42":
			_, _ = w.Write([]byte(`{"id":42,"state":"OPEN","destination":{"branch":{"name":"main"}}}`))
		case "/repositories/ws/repo/merge-queues":
			require.Equal(t, "main", r.URL.Query().Get("target_branch"))
			_, _ = w.Write([]byte(`{"values":[{"uuid":"q","state":"PAUSED"}]}`))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	err := cloud.NewClient(srv.Client(), srv.URL, "tok", "").EnableAutoMerge("ws", "repo", 42, "merge")
	require.Error(t, err)
	require.Zero(t, writes, "a missing active queue must never fall through to ordinary merge")
}

func TestMergePR_DoesNotReportQueuedPullRequestAsMerged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"id":42,"state":"OPEN","queued":true}`))
	}))
	defer srv.Close()
	_, err := cloud.NewClient(srv.Client(), srv.URL, "tok", "").MergePR("ws", "repo", 42, backend.MergePRInput{})
	require.Error(t, err)
}

func TestEnableAutoMerge_VerifiesOfficialQueueAfterEmptyAcceptedResponse(t *testing.T) {
	for _, tc := range []struct{ strategy, wire string }{{"merge", "merge_commit"}, {"squash", "squash"}, {"rebase", "rebase_merge"}} {
		t.Run(tc.strategy, func(t *testing.T) {
			checked := false
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/repositories/ws/repo/pullrequests/42":
					_, _ = w.Write([]byte(`{"id":42,"state":"OPEN","destination":{"branch":{"name":"main"}}}`))
				case "/repositories/ws/repo/merge-queues":
					_, _ = w.Write([]byte(`{"values":[{"uuid":"q","state":"ACTIVE"}]}`))
				case "/repositories/ws/repo/pullrequests/42/merge":
					var body map[string]string
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					require.Equal(t, tc.wire, body["merge_strategy"])
					w.WriteHeader(202)
				case "/repositories/ws/repo/merge-queues/q/queue/42":
					checked = true
					_, _ = w.Write([]byte(`{"status":"QUEUED","pullrequest":{"id":42}}`))
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer srv.Close()
			require.NoError(t, cloud.NewClient(srv.Client(), srv.URL, "tok", "").EnableAutoMerge("ws", "repo", 42, tc.strategy))
			require.True(t, checked)
		})
	}
}

func TestDisableAutoMerge_VerifiesRemovalFromDrainingQueueAndOpenPR(t *testing.T) {
	for _, tc := range []struct {
		name, finalState string
		wantError        bool
	}{{"cancelled", "OPEN", false}, {"merge won race", "MERGED", true}} {
		t.Run(tc.name, func(t *testing.T) {
			removed := false
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/repositories/ws/repo/pullrequests/42":
					state := "OPEN"
					if removed {
						state = tc.finalState
					}
					_, _ = fmt.Fprintf(w, `{"id":42,"state":%q,"queued":false,"destination":{"branch":{"name":"main"}}}`, state)
				case "/repositories/ws/repo/merge-queues":
					_, _ = w.Write([]byte(`{"values":[{"uuid":"q","state":"DRAINING"}]}`))
				case "/repositories/ws/repo/merge-queues/q/queue/42":
					if r.Method == http.MethodDelete {
						removed = true
						require.Empty(t, r.Header.Get("Content-Type"))
						w.WriteHeader(202)
						return
					}
					if removed {
						w.WriteHeader(404)
						return
					}
					_, _ = w.Write([]byte(`{"status":"QUEUED","pullrequest":{"id":42}}`))
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer srv.Close()
			err := cloud.NewClient(srv.Client(), srv.URL, "tok", "").DisableAutoMerge("ws", "repo", 42)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.True(t, removed)
		})
	}
}

func TestGetPR_MapsPublicQueuedField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":42,"state":"OPEN","queued":true}`))
	}))
	defer srv.Close()
	pr, err := cloud.NewClient(srv.Client(), srv.URL, "tok", "").GetPR("ws", "repo", 42)
	require.NoError(t, err)
	require.NotNil(t, pr.AutoMerge)
	require.True(t, pr.AutoMerge.Enabled)
	require.Empty(t, pr.AutoMerge.Strategy)
}

func TestDisableAutoMerge_PreservesAuthenticationFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repositories/ws/repo/pullrequests/42":
			_, _ = w.Write([]byte(`{"id":42,"state":"OPEN","destination":{"branch":{"name":"main"}}}`))
		case "/repositories/ws/repo/merge-queues":
			_, _ = w.Write([]byte(`{"values":[{"uuid":"q","state":"INACTIVE"}]}`))
		default:
			w.WriteHeader(401)
			_, _ = w.Write([]byte(`{"error":{"message":"token expired"}}`))
		}
	}))
	defer srv.Close()
	err := cloud.NewClient(srv.Client(), srv.URL, "tok", "").DisableAutoMerge("ws", "repo", 42)
	require.ErrorIs(t, err, backend.ErrAuth)
}

func TestEnableAutoMerge_AcceptsMergeCompletedBeforeMembershipRead(t *testing.T) {
	merged := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repositories/ws/repo/pullrequests/42":
			state := "OPEN"
			if merged {
				state = "MERGED"
			}
			_, _ = fmt.Fprintf(w, `{"id":42,"state":%q,"destination":{"branch":{"name":"main"}}}`, state)
		case "/repositories/ws/repo/merge-queues":
			_, _ = w.Write([]byte(`{"values":[{"uuid":"q","state":"ACTIVE"}]}`))
		case "/repositories/ws/repo/pullrequests/42/merge":
			merged = true
			w.WriteHeader(202)
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	require.NoError(t, cloud.NewClient(srv.Client(), srv.URL, "tok", "").EnableAutoMerge("ws", "repo", 42, "merge"))
}

func TestEnableAutoMerge_DoesNotConfirmRemovingOrMalformedMembership(t *testing.T) {
	for _, item := range []string{`{"status":"REMOVING","pullrequest":{"id":42}}`, `{"status":"QUEUED","pullrequest":{"id":99}}`, `{"status":"QUEUED"}`, `{"status":"QUEUED","pullrequest":{"id":42},"merge_queue":{"uuid":"other"}}`} {
		t.Run(item, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/repositories/ws/repo/pullrequests/42":
					_, _ = w.Write([]byte(`{"id":42,"state":"OPEN","destination":{"branch":{"name":"main"}}}`))
				case "/repositories/ws/repo/merge-queues":
					_, _ = w.Write([]byte(`{"values":[{"uuid":"q","state":"ACTIVE"}]}`))
				case "/repositories/ws/repo/pullrequests/42/merge":
					w.WriteHeader(202)
				default:
					_, _ = w.Write([]byte(item))
				}
			}))
			defer srv.Close()
			err := cloud.NewClient(srv.Client(), srv.URL, "tok", "").EnableAutoMerge("ws", "repo", 42, "merge")
			require.Error(t, err)
			var de *backend.DomainError
			require.ErrorAs(t, err, &de)
			require.Equal(t, backend.CodePRMergeQueueUnconfirmed, de.Code)
		})
	}
}
