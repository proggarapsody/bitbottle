package testhelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// StubResponse describes a single canned HTTP response for NewTestServer.
// If Handler is set it takes precedence over Status/Body.
type StubResponse struct {
	Method     string
	PathSuffix string
	Status     int
	Body       any
	Handler    http.HandlerFunc
}

// NewTestServer returns an httptest.NewTLSServer that dispatches requests to
// the first stub whose Method + PathSuffix match. Unmatched requests fail the
// test immediately.
func NewTestServer(t *testing.T, stubs ...StubResponse) *httptest.Server {
	t.Helper()
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
					t.Errorf("testhelpers: encoding stub body: %v", err)
				}
			}
			return
		}
		t.Errorf("testhelpers: unmatched request %s %s", r.Method, r.URL.Path)
		http.Error(w, "unmatched request", http.StatusNotImplemented)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// BitbucketServer is a fluent builder around NewTestServer that knows common
// Bitbucket Data Center endpoints.
type BitbucketServer struct {
	t     *testing.T
	stubs []StubResponse
}

// NewBitbucketServer creates a new fluent BitbucketServer builder.
func NewBitbucketServer(t *testing.T) *BitbucketServer {
	t.Helper()
	return &BitbucketServer{t: t}
}

// WithRepoList stubs GET /rest/api/1.0/repos with the provided list of repos.
func (b *BitbucketServer) WithRepoList(repos []map[string]any) *BitbucketServer {
	values := make([]any, len(repos))
	for i, r := range repos {
		values[i] = r
	}
	b.stubs = append(b.stubs, StubResponse{
		Method:     http.MethodGet,
		PathSuffix: "/rest/api/1.0/repos",
		Status:     http.StatusOK,
		Body:       PagedResponse(values),
	})
	return b
}

// WithRepo stubs GET /rest/api/1.0/projects/{project}/repos/{slug}.
func (b *BitbucketServer) WithRepo(project, slug string, repo map[string]any) *BitbucketServer {
	b.stubs = append(b.stubs, StubResponse{
		Method:     http.MethodGet,
		PathSuffix: fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s", project, slug),
		Status:     http.StatusOK,
		Body:       repo,
	})
	return b
}

// WithPRList stubs GET /rest/api/1.0/projects/{project}/repos/{slug}/pull-requests.
func (b *BitbucketServer) WithPRList(project, slug string, prs []map[string]any) *BitbucketServer {
	values := make([]any, len(prs))
	for i, p := range prs {
		values[i] = p
	}
	b.stubs = append(b.stubs, StubResponse{
		Method:     http.MethodGet,
		PathSuffix: fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests", project, slug),
		Status:     http.StatusOK,
		Body:       PagedResponse(values),
	})
	return b
}

// WithPR stubs GET .../pull-requests/{id}.
func (b *BitbucketServer) WithPR(project, slug string, id int, pr map[string]any) *BitbucketServer {
	b.stubs = append(b.stubs, StubResponse{
		Method:     http.MethodGet,
		PathSuffix: fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests/%d", project, slug, id),
		Status:     http.StatusOK,
		Body:       pr,
	})
	return b
}

// WithApplicationProperties stubs GET /rest/api/1.0/application-properties.
func (b *BitbucketServer) WithApplicationProperties(version string) *BitbucketServer {
	b.stubs = append(b.stubs, StubResponse{
		Method:     http.MethodGet,
		PathSuffix: "/rest/api/1.0/application-properties",
		Status:     http.StatusOK,
		Body: map[string]any{
			"version":     version,
			"buildNumber": "1",
			"displayName": "Bitbucket",
		},
	})
	return b
}

// WithCurrentUser stubs the "current user" endpoint used for whoami-style calls.
func (b *BitbucketServer) WithCurrentUser(user map[string]any) *BitbucketServer {
	b.stubs = append(b.stubs, StubResponse{
		Method:     http.MethodGet,
		PathSuffix: "/plugins/servlet/applinks/whoami",
		Status:     http.StatusOK,
		Body:       user,
	})
	b.stubs = append(b.stubs, StubResponse{
		Method:     http.MethodGet,
		PathSuffix: "/rest/api/1.0/users/" + fmt.Sprint(user["slug"]),
		Status:     http.StatusOK,
		Body:       user,
	})
	return b
}

// WithError stubs an error response matching any request whose path ends with
// pathSuffix.
func (b *BitbucketServer) WithError(pathSuffix string, status int, message string) *BitbucketServer {
	b.stubs = append(b.stubs, StubResponse{
		PathSuffix: pathSuffix,
		Status:     status,
		Body: map[string]any{
			"errors": []map[string]any{
				{"message": message},
			},
		},
	})
	return b
}

// Build constructs the httptest.Server using the accumulated stubs.
func (b *BitbucketServer) Build() *httptest.Server {
	return NewTestServer(b.t, b.stubs...)
}

// PagedResponse wraps the given values slice in a Bitbucket paged envelope.
func PagedResponse(values any) map[string]any {
	size := 0
	switch v := values.(type) {
	case []any:
		size = len(v)
	case []map[string]any:
		size = len(v)
	}
	return map[string]any{
		"values":     values,
		"size":       size,
		"isLastPage": true,
		"start":      0,
		"limit":      25,
	}
}
