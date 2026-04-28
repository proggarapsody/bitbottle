package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/api"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

var jsonUnmarshal = json.Unmarshal

func iostreamsWithStdin(data string) *iostreams.IOStreams {
	ios := iostreams.Test()
	ios.In = io.NopCloser(strings.NewReader(data))
	ios.Out = &bytes.Buffer{}
	ios.ErrOut = &bytes.Buffer{}
	return ios
}

const apiHostConfig = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

// stubHTTP records the last request and returns a canned response.
// For multi-request tests (pagination), set Sequence: each call peels one entry.
type stubHTTP struct {
	req      *http.Request
	requests []*http.Request
	reqBody  []byte
	status   int
	respBody string
	header   http.Header
	sequence []string
	calls    int
}

func (s *stubHTTP) Do(req *http.Request) (*http.Response, error) {
	s.req = req
	s.requests = append(s.requests, req)
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		s.reqBody = body
	}
	if s.status == 0 {
		s.status = 200
	}
	if s.header == nil {
		s.header = http.Header{"Content-Type": []string{"application/json"}}
	}
	respBody := s.respBody
	if len(s.sequence) > 0 {
		idx := s.calls
		if idx >= len(s.sequence) {
			idx = len(s.sequence) - 1
		}
		respBody = s.sequence[idx]
	}
	s.calls++
	return &http.Response{
		StatusCode: s.status,
		Body:       io.NopCloser(strings.NewReader(respBody)),
		Header:     s.header,
	}, nil
}

func TestAPI_Method_OverridesVerb(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: ``}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"--method", "DELETE", "2.0/repositories/me/x"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "DELETE", stub.req.Method)
}

func TestAPI_HostnameFlag_TargetsHost(t *testing.T) {
	t.Parallel()

	const twoHosts = "" +
		"bb.example.com:\n  oauth_token: tok1\n  user: alice\n  git_protocol: ssh\n" +
		"bb.other.com:\n  oauth_token: tok2\n  user: bob\n  git_protocol: ssh\n"

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: twoHosts,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"--hostname", "bb.other.com", "rest/api/1.0/projects"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "https://bb.other.com/rest/api/1.0/projects", stub.req.URL.String())
	assert.Equal(t, "Bearer tok2", stub.req.Header.Get("Authorization"))
}

func TestAPI_AmbiguousHost_Errors(t *testing.T) {
	t.Parallel()

	const twoHosts = "" +
		"bb.example.com:\n  oauth_token: tok1\n  user: alice\n  git_protocol: ssh\n" +
		"bb.other.com:\n  oauth_token: tok2\n  user: bob\n  git_protocol: ssh\n"

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: twoHosts,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"2.0/user"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple hosts")
}

func TestAPI_Header_AddsCustomHeader(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"-H", "X-Custom: yes", "-H", "X-Another: no", "2.0/user"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "yes", stub.req.Header.Get("X-Custom"))
	assert.Equal(t, "no", stub.req.Header.Get("X-Another"))
}

func TestAPI_Non2xx_ReturnsError(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{status: 404, respBody: `{"error":{"message":"not found"}}`}
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"2.0/missing"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
	// Body still printed so user can see API error envelope.
	assert.Contains(t, out.String(), "not found")
}

func TestAPI_TypedField_BuildsJSONBodyAndDefaultsToPOST(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{
		"-F", "title=My PR",
		"-F", "draft=true",
		"-F", "id=42",
		"2.0/repositories/me/x/pullrequests",
	})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "POST", stub.req.Method)
	assert.Equal(t, "application/json", stub.req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"title":"My PR","draft":true,"id":42}`, string(stub.reqBody))
}

func TestAPI_StringField_KeepsValueAsString(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"-f", "version=42", "2.0/repositories/me/x"})
	require.NoError(t, cmd.Execute())

	assert.JSONEq(t, `{"version":"42"}`, string(stub.reqBody))
}

func TestAPI_Input_StreamsRawBody(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
		IOStreams:     iostreamsWithStdin(`{"raw":"yes"}`),
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"--input", "-", "-X", "PUT", "2.0/repositories/me/x"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "PUT", stub.req.Method)
	assert.Equal(t, `{"raw":"yes"}`, string(stub.reqBody))
}

func TestAPI_JQ_FiltersJSONResponse(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{"values":[{"name":"a"},{"name":"b"}]}`}
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"--jq", ".values[].name", "2.0/repositories/me"})
	require.NoError(t, cmd.Execute())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Equal(t, []string{`"a"`, `"b"`}, lines)
}

func TestAPI_JQ_NonJSONResponse_Errors(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `not json`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"--jq", ".foo", "2.0/x"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jq")
}

func TestAPI_VariableExpansion_FromBaseRepo(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{Host: "bb.example.com", Project: "PROJ", Slug: "myrepo"}, nil
	}

	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"rest/api/1.0/projects/{project}/repos/{slug}/pull-requests"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t,
		"https://bb.example.com/rest/api/1.0/projects/PROJ/repos/myrepo/pull-requests",
		stub.req.URL.String(),
	)
}

func TestAPI_VariableExpansion_CloudAliases(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{Host: "bitbucket.org", Project: "myws", Slug: "myrepo"}, nil
	}

	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"2.0/repositories/{workspace}/{repo_slug}/pullrequests"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t,
		"https://bb.example.com/2.0/repositories/myws/myrepo/pullrequests",
		stub.req.URL.String(),
	)
}

func TestAPI_VariableExpansion_NoRepo_Errors(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{}`}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	f.BaseRepo = func() (bbrepo.RepoRef, error) {
		return bbrepo.RepoRef{}, fmt.Errorf("no remotes")
	}

	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"2.0/repositories/{workspace}/x"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "{workspace}")
}

func TestAPI_Paginate_CloudFollowsNextURL(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{
		sequence: []string{
			`{"values":[{"id":1},{"id":2}],"next":"https://bb.example.com/2.0/repositories/me/x?page=2"}`,
			`{"values":[{"id":3}]}`,
		},
	}
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"--paginate", "2.0/repositories/me/x"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 2, stub.calls)
	assert.Equal(t, "https://bb.example.com/2.0/repositories/me/x?page=2", stub.requests[1].URL.String())

	// Output: merged JSON array of all `values`.
	var got []map[string]any
	require.NoError(t, jsonUnmarshal(out.Bytes(), &got))
	assert.Len(t, got, 3)
}

func TestAPI_Paginate_ServerWalksNextPageStart(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{
		sequence: []string{
			`{"values":[{"id":1}],"isLastPage":false,"nextPageStart":25}`,
			`{"values":[{"id":2}],"isLastPage":true}`,
		},
	}
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})
	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"--paginate", "rest/api/1.0/projects/PROJ/repos/x/pull-requests"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 2, stub.calls)
	assert.Equal(t, "25", stub.requests[1].URL.Query().Get("start"))

	var got []map[string]any
	require.NoError(t, jsonUnmarshal(out.Bytes(), &got))
	assert.Len(t, got, 2)
}

func TestAPI_GET_PrintsBody(t *testing.T) {
	t.Parallel()

	stub := &stubHTTP{respBody: `{"username":"alice"}`}
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: apiHostConfig,
		HTTPClient:    stub,
	})

	cmd := api.NewCmdAPI(f)
	cmd.SetArgs([]string{"2.0/user"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, stub.req)
	assert.Equal(t, "GET", stub.req.Method)
	assert.Equal(t, "https://bb.example.com/2.0/user", stub.req.URL.String())
	assert.Equal(t, "Bearer tok", stub.req.Header.Get("Authorization"))
	assert.Equal(t, `{"username":"alice"}`, strings.TrimSpace(out.String()))
}
