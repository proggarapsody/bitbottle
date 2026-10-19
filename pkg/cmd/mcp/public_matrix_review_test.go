package mcp_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	cmdmcp "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
	"github.com/proggarapsody/bitbottle/test/qamatrix"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const reviewChildEnv = "BITBOTTLE_TEST_MCP_REVIEW_CHILD"

// The wrappers expose exactly the optional read capabilities present on both
// real adapters. FakeClient itself exposes unrelated optional capabilities.
type reviewCloudClient struct {
	backend.Client
	backend.CommitFileClient
	backend.PRCommitClient
}
type reviewServerClient struct {
	backend.Client
	backend.CommitFileClient
	backend.PRCommitClient
	taskState *reviewTaskState
}

type reviewTaskState struct{ comments map[int]backend.PRComment }

func (c reviewServerClient) SetPRCommentState(ns, slug string, prID, taskID int, state string) error {
	if ns != matrixProject || slug != matrixRepo || prID != 42 {
		return fmt.Errorf("unexpected task target %s/%s #%d", ns, slug, prID)
	}
	if taskID == 99 {
		return reviewDenied(matrixServerHost, "comment", "99")
	}
	if taskID != 13 || (state != "RESOLVED" && state != "OPEN") {
		return fmt.Errorf("unexpected task state %d %q", taskID, state)
	}
	task, ok := c.taskState.comments[taskID]
	if !ok {
		return fmt.Errorf("missing task %d", taskID)
	}
	task.State = state
	c.taskState.comments[taskID] = task
	return nil
}

func (c reviewServerClient) ListCommentReactions(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
	if ns != matrixProject || slug != matrixRepo || prID != 42 {
		return nil, fmt.Errorf("unexpected reaction target %s/%s #%d", ns, slug, prID)
	}
	if commentID == 11 {
		return []backend.CommentReaction{{Emoji: "thumbs_up", Users: []backend.User{{Slug: "alice"}}}}, nil
	}
	if commentID == 13 || commentID == 14 {
		return []backend.CommentReaction{}, nil
	}
	return nil, fmt.Errorf("unexpected reaction comment %d", commentID)
}

func (c reviewServerClient) AddCommentReaction(ns, slug string, prID, commentID int, emoji string) error {
	return fmt.Errorf("unexpected reaction write")
}

func (c reviewServerClient) RemoveCommentReaction(ns, slug string, prID, commentID int, emoji string) error {
	return fmt.Errorf("unexpected reaction delete")
}

func TestMCPReviewChild(t *testing.T) {
	if os.Getenv(reviewChildEnv) != "1" {
		return
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{ConfigDir: os.Getenv("BITBOTTLE_TEST_MCP_CONFIG_DIR"), InitialConfig: matrixCloudHost + ":\n  oauth_token: test-token\n" + matrixServerHost + ":\n  oauth_token: test-token\n"})
	cloudState := &reviewTaskState{comments: map[int]backend.PRComment{11: {ID: 11, Text: "Original note"}}}
	serverState := &reviewTaskState{comments: map[int]backend.PRComment{11: {ID: 11, Text: "Original note"}}}
	cloud := newReviewFixture(t, matrixCloudHost, cloudState)
	server := newReviewFixture(t, matrixServerHost, serverState)
	f.Backend = func(host string) (backend.Client, error) {
		switch host {
		case matrixCloudHost:
			return reviewCloudClient{Client: cloud, CommitFileClient: cloud, PRCommitClient: cloud}, nil
		case matrixServerHost:
			return reviewServerClient{Client: server, CommitFileClient: server, PRCommitClient: server, taskState: serverState}, nil
		}
		return nil, fmt.Errorf("unexpected host %q", host)
	}
	cmd := cmdmcp.NewCmdMCP(f)
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func reviewDenied(host, resource, id string) error {
	return &backend.DomainError{Kind: backend.ErrPermission, Code: backend.CodePermWriteRequired, Host: host, Resource: resource, ID: id, Message: "fixture denies access"}
}

func newReviewFixture(t *testing.T, host string, taskState *reviewTaskState) *testhelpers.FakeClient {
	c := &testhelpers.FakeClient{T: t}
	knownCommit := backend.Commit{Hash: "abc123", Message: "Fix sync race"}
	knownFiles := []backend.DiffStatEntry{{Path: "src/sync.go", Status: "modified", Additions: 3, Deletions: 1}}
	tags := map[string]backend.Tag{"v1.0.0": {Name: "v1.0.0", Hash: "abc123"}}
	comments := taskState.comments
	checkRepo := func(ns, slug string) error {
		if ns != matrixProject {
			return fmt.Errorf("unexpected project %q", ns)
		}
		if slug == "denied" {
			return reviewDenied(host, "repository", slug)
		}
		if slug != matrixRepo {
			return fmt.Errorf("unexpected slug %q", slug)
		}
		return nil
	}
	checkPR := func(ns, slug string, id int) error {
		if err := checkRepo(ns, slug); err != nil {
			return err
		}
		if id == 404 {
			return reviewDenied(host, "pull-request", "404")
		}
		if id != 42 {
			return fmt.Errorf("unexpected PR %d", id)
		}
		return nil
	}
	c.ListCommitsFn = func(ns, slug, branch string, limit int) ([]backend.Commit, error) {
		if err := checkRepo(ns, slug); err != nil {
			return nil, err
		}
		if branch != "release/1.0" || limit != 4 {
			return nil, fmt.Errorf("unexpected list commits %q %d", branch, limit)
		}
		return []backend.Commit{knownCommit}, nil
	}
	c.GetCommitFn = func(ns, slug, hash string) (backend.Commit, error) {
		if err := checkRepo(ns, slug); err != nil {
			return backend.Commit{}, err
		}
		if hash == "denied" {
			return backend.Commit{}, reviewDenied(host, "commit", hash)
		}
		if hash != "abc123" {
			return backend.Commit{}, fmt.Errorf("unexpected hash %q", hash)
		}
		return knownCommit, nil
	}
	c.ListCommitFilesFn = func(ns, slug, hash string) ([]backend.DiffStatEntry, error) {
		if err := checkRepo(ns, slug); err != nil {
			return nil, err
		}
		if hash == "denied" {
			return nil, reviewDenied(host, "commit", hash)
		}
		if hash != "abc123" {
			return nil, fmt.Errorf("unexpected hash %q", hash)
		}
		return knownFiles, nil
	}
	c.ListPRCommitsFn = func(ns, slug string, id int) ([]backend.Commit, error) {
		if err := checkPR(ns, slug, id); err != nil {
			return nil, err
		}
		return []backend.Commit{knownCommit}, nil
	}
	c.ListTagsFn = func(ns, slug string, limit int) ([]backend.Tag, error) {
		if err := checkRepo(ns, slug); err != nil {
			return nil, err
		}
		if limit != 5 {
			return nil, fmt.Errorf("unexpected tag limit %d", limit)
		}
		out := []backend.Tag{tags["v1.0.0"]}
		if tag, ok := tags["v1.1.0"]; ok {
			out = append(out, tag)
		}
		return out, nil
	}
	c.CreateTagFn = func(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
		if err := checkRepo(ns, slug); err != nil {
			return backend.Tag{}, err
		}
		if in.Name == "denied" {
			return backend.Tag{}, reviewDenied(host, "tag", in.Name)
		}
		if in != (backend.CreateTagInput{Name: "v1.1.0", StartAt: "abc123", Message: "Release 1.1"}) {
			return backend.Tag{}, fmt.Errorf("unexpected tag input %+v", in)
		}
		tag := backend.Tag{Name: in.Name, Hash: "abc123", Message: in.Message}
		tags[in.Name] = tag
		return tag, nil
	}
	c.DeleteTagFn = func(ns, slug, name string) error {
		if err := checkRepo(ns, slug); err != nil {
			return err
		}
		if name == "denied" {
			return reviewDenied(host, "tag", name)
		}
		if name != "v1.1.0" {
			return fmt.Errorf("unexpected tag %q", name)
		}
		delete(tags, name)
		return nil
	}
	c.ListPRCommentsFn = func(ns, slug string, id int) ([]backend.PRComment, error) {
		if err := checkPR(ns, slug, id); err != nil {
			return nil, err
		}
		out := []backend.PRComment{comments[11]}
		if comment, ok := comments[12]; ok {
			out = append(out, comment)
		}
		if task, ok := comments[13]; ok {
			out = append(out, task)
		}
		if task, ok := comments[14]; ok {
			out = append(out, task)
		}
		if comment, ok := comments[21]; ok {
			out = append(out, comment)
		}
		if comment, ok := comments[22]; ok {
			out = append(out, comment)
		}
		return out, nil
	}
	c.AddPRCommentFn = func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
		if err := checkPR(ns, slug, id); err != nil {
			return backend.PRComment{}, err
		}
		if in.Text == "denied" {
			return backend.PRComment{}, reviewDenied(host, "comment", "new")
		}
		if in.Text == "Fix this" && in.Severity == "BLOCKER" {
			task := backend.PRComment{ID: 13, Text: in.Text, Severity: "BLOCKER", State: "OPEN"}
			comments[13] = task
			return task, nil
		}
		if in.Text == "Reply task" && in.Severity == "BLOCKER" && in.Parent != nil && *in.Parent == 11 {
			task := backend.PRComment{ID: 14, Text: in.Text, Severity: "BLOCKER", State: "OPEN", ParentID: 11}
			comments[14] = task
			return task, nil
		}
		if in.Text == "Range note" {
			startLine := 0
			if host == matrixCloudHost {
				startLine = 16
			}
			want := backend.AddPRCommentInput{Text: "Range note", Inline: &backend.PRCommentInline{Path: "src/sync.go", Side: "old", Line: 18, StartLine: startLine}}
			if !reflect.DeepEqual(in, want) {
				return backend.PRComment{}, fmt.Errorf("unexpected range input %+v", in)
			}
			comment := backend.PRComment{ID: 21, Text: in.Text, Inline: in.Inline}
			comments[21] = comment
			return comment, nil
		}
		if in.Text == "Thread reply" {
			if in.Severity != "" || in.Inline != nil || in.Parent == nil || *in.Parent != 11 {
				return backend.PRComment{}, fmt.Errorf("unexpected reply input %+v", in)
			}
			comment := backend.PRComment{ID: 22, Text: in.Text, ParentID: 11}
			comments[22] = comment
			return comment, nil
		}
		want := backend.AddPRCommentInput{Text: "Please revise", Inline: &backend.PRCommentInline{Path: "src/sync.go", Side: "new", Line: 18}}
		if !reflect.DeepEqual(in, want) {
			return backend.PRComment{}, fmt.Errorf("unexpected comment input %+v", in)
		}
		comment := backend.PRComment{ID: 12, Text: in.Text, Inline: in.Inline}
		comments[12] = comment
		return comment, nil
	}
	c.EditPRCommentFn = func(ns, slug string, id, commentID int, body string) (backend.PRComment, error) {
		if err := checkPR(ns, slug, id); err != nil {
			return backend.PRComment{}, err
		}
		if commentID == 99 {
			return backend.PRComment{}, reviewDenied(host, "comment", "99")
		}
		if commentID != 12 || body != "Please revise tests" {
			return backend.PRComment{}, fmt.Errorf("unexpected edit %d %q", commentID, body)
		}
		comment := comments[12]
		comment.Text = body
		comments[12] = comment
		return comment, nil
	}
	c.DeletePRCommentFn = func(ns, slug string, id, commentID int) error {
		if err := checkPR(ns, slug, id); err != nil {
			return err
		}
		if commentID == 99 {
			return reviewDenied(host, "comment", "99")
		}
		if commentID != 12 {
			return fmt.Errorf("unexpected delete %d", commentID)
		}
		delete(comments, 12)
		return nil
	}
	return c
}

func startReviewSession(t *testing.T) *stdioSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	t.Cleanup(cancel)
	configDir := t.TempDir()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMCPReviewChild$")
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + configDir, "TMPDIR=" + configDir, reviewChildEnv + "=1", "BITBOTTLE_TEST_MCP_CONFIG_DIR=" + configDir}
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())
	s := &stdioSession{ctx: ctx, cmd: cmd, stdin: stdin, relayDone: make(chan error, 1), rawStdout: &bytes.Buffer{}, stderr: &stderr}
	t.Cleanup(func() {
		_ = stdin.Close()
		if !s.waited {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	input, output := io.Pipe()
	s.clientInput = input
	s.relayOutput = output
	go func() {
		_, copyErr := io.Copy(io.MultiWriter(output, s.rawStdout), stdout)
		_ = output.CloseWithError(copyErr)
		s.relayDone <- copyErr
	}()
	protocol := client.NewClient(transport.NewIO(input, stdin, io.NopCloser(strings.NewReader(""))))
	s.client = protocol
	t.Cleanup(func() { _ = protocol.Close(); _ = input.Close(); _ = output.Close() })
	require.NoError(t, protocol.Start(ctx))
	s.initialized, err = protocol.Initialize(ctx, mcplib.InitializeRequest{Params: mcplib.InitializeParams{ProtocolVersion: mcplib.LATEST_PROTOCOL_VERSION, ClientInfo: mcplib.Implementation{Name: "bitbottle-review-matrix-test", Version: "1"}}})
	require.NoError(t, err)
	return s
}

func reviewRecord(t *testing.T, name, host, outcome string, parameters ...string) {
	t.Helper()
	provider := "cloud"
	if host == matrixServerHost {
		provider = "server"
	}
	assertions := []string{"output", "input", "effect"}
	if outcome == "error" || outcome == "unsupported" {
		assertions = []string{"error", "input", "effect"}
	}
	qamatrix.Record(t, qamatrix.Scenario{Operation: "mcp:" + name, Backend: provider, Outcome: outcome, Assertions: assertions, Flags: parameters})
}

func TestMCPPublicReviewMatrix(t *testing.T) {
	s := startReviewSession(t)
	listed, err := s.client.ListTools(s.ctx, mcplib.ListToolsRequest{})
	require.NoError(t, err)
	tools := make(map[string]mcplib.Tool, len(listed.Tools))
	for _, tool := range listed.Tools {
		tools[tool.Name] = tool
	}
	for _, name := range []string{"list_commits", "get_commit", "list_commit_files", "list_pr_commits", "list_pr_comments", "add_pr_comment", "edit_pr_comment", "delete_pr_comment", "list_tags", "create_tag", "delete_tag", "list_pr_tasks", "create_pr_task", "resolve_pr_task", "reopen_pr_task"} {
		require.Contains(t, tools, name)
	}
	for _, host := range []string{matrixCloudHost, matrixServerHost} {
		t.Run(host, func(t *testing.T) { reviewProviderScenario(t, s, host) })
	}
	cloudArgs := matrixArgs(matrixCloudHost)
	before := matrixCallOK[[]backend.PRComment](t, s, "list_pr_comments", matrixWith(cloudArgs, "id", 42))
	matrixCallError(t, s, "create_pr_task", matrixWith(matrixWith(cloudArgs, "pr_id", 42), "body", "Fix this"), "host.unsupported", matrixCloudHost)
	require.Equal(t, before, matrixCallOK[[]backend.PRComment](t, s, "list_pr_comments", matrixWith(cloudArgs, "id", 42)))
	reviewRecord(t, "create_pr_task", matrixCloudHost, "unsupported", "project", "slug", "pr_id", "body")
	s.close(t, len(listed.Tools))
}

func reviewProviderScenario(t *testing.T, s *stdioSession, host string) {
	base := matrixArgs(host)
	commit := backend.Commit{Hash: "abc123", Message: "Fix sync race"}
	files := []backend.DiffStatEntry{{Path: "src/sync.go", Status: "modified", Additions: 3, Deletions: 1}}
	listTags := func() []backend.Tag {
		return matrixCallOK[[]backend.Tag](t, s, "list_tags", matrixWith(base, "limit", 5))
	}
	listComments := func() []backend.PRComment {
		return matrixCallOK[[]backend.PRComment](t, s, "list_pr_comments", matrixWith(base, "id", 42))
	}
	listCommits := matrixCallOK[[]backend.Commit](t, s, "list_commits", matrixWith(matrixWith(base, "branch", "release/1.0"), "limit", 4))
	require.Equal(t, []backend.Commit{commit}, listCommits)
	reviewRecord(t, "list_commits", host, "success", "project", "slug", "branch", "limit")
	matrixCallError(t, s, "list_commits", matrixWith(matrixWith(matrixWith(base, "slug", "denied"), "branch", "release/1.0"), "limit", 4), "perm.write_required", host, "denied")
	reviewRecord(t, "list_commits", host, "error", "project", "slug")
	require.Equal(t, commit, matrixCallOK[backend.Commit](t, s, "get_commit", matrixWith(base, "hash", "abc123")))
	reviewRecord(t, "get_commit", host, "success", "project", "slug", "hash")
	matrixCallError(t, s, "get_commit", matrixWith(base, "hash", "denied"), "perm.write_required", host, "denied")
	reviewRecord(t, "get_commit", host, "error", "hash")
	require.Equal(t, files, matrixCallOK[[]backend.DiffStatEntry](t, s, "list_commit_files", map[string]any{"hostname": host, "repo": "PROJ/repo", "hash": "abc123"}))
	reviewRecord(t, "list_commit_files", host, "success", "repo", "hash")
	matrixCallError(t, s, "list_commit_files", map[string]any{"hostname": host, "repo": "PROJ/repo", "hash": "denied"}, "perm.write_required", host, "denied")
	reviewRecord(t, "list_commit_files", host, "error", "repo", "hash")
	require.Equal(t, []backend.Commit{commit}, matrixCallOK[[]backend.Commit](t, s, "list_pr_commits", matrixWith(base, "pr_id", 42)))
	reviewRecord(t, "list_pr_commits", host, "success", "project", "slug", "pr_id")
	matrixCallError(t, s, "list_pr_commits", matrixWith(base, "pr_id", 404), "perm.write_required", host, "404")
	reviewRecord(t, "list_pr_commits", host, "error", "pr_id")
	require.Equal(t, []backend.Tag{{Name: "v1.0.0", Hash: "abc123"}}, listTags())
	reviewRecord(t, "list_tags", host, "success", "project", "slug", "limit")
	matrixCallError(t, s, "list_tags", matrixWith(matrixWith(base, "slug", "denied"), "limit", 5), "perm.write_required", host, "denied")
	reviewRecord(t, "list_tags", host, "error", "slug", "limit")
	tag := matrixCallOK[backend.Tag](t, s, "create_tag", matrixWith(matrixWith(matrixWith(base, "name", "v1.1.0"), "start_at", "abc123"), "message", "Release 1.1"))
	require.Equal(t, backend.Tag{Name: "v1.1.0", Hash: "abc123", Message: "Release 1.1"}, tag)
	require.Contains(t, listTags(), tag)
	reviewRecord(t, "create_tag", host, "success", "project", "slug", "name", "start_at", "message")
	matrixCallError(t, s, "create_tag", matrixWith(matrixWith(base, "name", "denied"), "start_at", "abc123"), "perm.write_required", host, "denied")
	require.Contains(t, listTags(), tag)
	reviewRecord(t, "create_tag", host, "error", "name", "start_at")
	require.Empty(t, matrixCallOK[map[string]any](t, s, "delete_tag", matrixWith(base, "name", "v1.1.0")))
	require.NotContains(t, listTags(), tag)
	reviewRecord(t, "delete_tag", host, "success", "project", "slug", "name")
	matrixCallError(t, s, "delete_tag", matrixWith(base, "name", "denied"), "perm.write_required", host, "denied")
	require.Equal(t, []backend.Tag{{Name: "v1.0.0", Hash: "abc123"}}, listTags())
	reviewRecord(t, "delete_tag", host, "error", "name")
	original := backend.PRComment{ID: 11, Text: "Original note"}
	require.Equal(t, []backend.PRComment{original}, listComments())
	reviewRecord(t, "list_pr_comments", host, "success", "project", "slug", "id")
	matrixCallError(t, s, "list_pr_comments", matrixWith(base, "id", 404), "perm.write_required", host, "404")
	reviewRecord(t, "list_pr_comments", host, "error", "id")
	inline := &backend.PRCommentInline{Path: "src/sync.go", Side: "new", Line: 18}
	added := matrixCallOK[backend.PRComment](t, s, "add_pr_comment", matrixWith(matrixWith(matrixWith(matrixWith(base, "id", 42), "body", "Please revise"), "inline_path", "src/sync.go"), "inline_line", 18))
	require.Equal(t, backend.PRComment{ID: 12, Text: "Please revise", Inline: inline}, added)
	require.Equal(t, []backend.PRComment{original, added}, listComments())
	require.Equal(t, []backend.PRComment{added}, matrixCallOK[[]backend.PRComment](t, s, "list_pr_comments", matrixWith(matrixWith(base, "id", 42), "inline_only", true)))
	reviewRecord(t, "add_pr_comment", host, "success", "project", "slug", "id", "body", "inline_path", "inline_line")
	reviewRecord(t, "list_pr_comments", host, "success", "inline_only")
	matrixCallError(t, s, "add_pr_comment", matrixWith(matrixWith(base, "id", 42), "body", "denied"), "perm.write_required", host, "new")
	require.Equal(t, []backend.PRComment{original, added}, listComments())
	reviewRecord(t, "add_pr_comment", host, "error", "id", "body")
	edited := matrixCallOK[backend.PRComment](t, s, "edit_pr_comment", matrixWith(matrixWith(matrixWith(base, "id", 42), "comment_id", 12), "body", "Please revise tests"))
	require.Equal(t, "Please revise tests", edited.Text)
	require.Equal(t, []backend.PRComment{original, edited}, listComments())
	reviewRecord(t, "edit_pr_comment", host, "success", "project", "slug", "id", "comment_id", "body")
	matrixCallError(t, s, "edit_pr_comment", matrixWith(matrixWith(matrixWith(base, "id", 42), "comment_id", 99), "body", "denied"), "perm.write_required", host, "99")
	require.Equal(t, []backend.PRComment{original, edited}, listComments())
	reviewRecord(t, "edit_pr_comment", host, "error", "id", "comment_id", "body")
	require.Empty(t, matrixCallOK[map[string]any](t, s, "delete_pr_comment", matrixWith(matrixWith(base, "id", 42), "comment_id", 12)))
	require.Equal(t, []backend.PRComment{original}, listComments())
	reviewRecord(t, "delete_pr_comment", host, "success", "project", "slug", "id", "comment_id")
	matrixCallError(t, s, "delete_pr_comment", matrixWith(matrixWith(base, "id", 42), "comment_id", 99), "perm.write_required", host, "99")
	require.Equal(t, []backend.PRComment{original}, listComments())
	reviewRecord(t, "delete_pr_comment", host, "error", "id", "comment_id")

	taskArgs := matrixWith(base, "pr_id", 42)
	listTasks := func(state string) []backend.PRComment {
		return matrixCallOK[[]backend.PRComment](t, s, "list_pr_tasks", matrixWith(taskArgs, "state", state))
	}
	if host == matrixCloudHost {
		// The current Cloud adapter represents ordinary comments, not task state.
		openComments := listTasks("open")
		require.Equal(t, []backend.PRComment{original}, openComments)
		require.Equal(t, openComments, listTasks("resolved"), "Cloud fallback leaves ordinary comments unfiltered by task state")
		require.Equal(t, []backend.PRComment{original}, listComments())
		reviewRecord(t, "list_pr_tasks", host, "success", "project", "slug", "pr_id", "state")
		matrixCallError(t, s, "list_pr_tasks", matrixWith(base, "pr_id", 404), "perm.write_required", host, "404")
		reviewRecord(t, "list_pr_tasks", host, "error", "pr_id")
		for _, name := range []string{"resolve_pr_task", "reopen_pr_task"} {
			matrixCallError(t, s, name, matrixWith(taskArgs, "task_id", 13), "host.unsupported", host)
			require.Equal(t, []backend.PRComment{original}, listComments())
			reviewRecord(t, name, host, "unsupported", "project", "slug", "pr_id", "task_id")
		}
		reviewOptionalCommentInputs(t, s, host, base)
		return
	}
	// An ordinary Server comment must not become a task merely because no
	// BLOCKER comment exists yet.
	beforeTask, err := s.call("list_pr_tasks", matrixWith(taskArgs, "state", "open"))
	require.NoError(t, err)
	require.False(t, beforeTask.IsError)
	require.JSONEq(t, "[]", resultText(t, beforeTask))
	require.Equal(t, []backend.PRComment{original}, listComments())
	reviewRecord(t, "list_pr_tasks", host, "success", "project", "slug", "pr_id", "state")
	task := matrixCallOK[map[string]any](t, s, "create_pr_task", matrixWith(taskArgs, "body", "Fix this"))
	require.Equal(t, float64(13), task["id"])
	require.Equal(t, "Fix this", task["text"])
	require.Equal(t, "BLOCKER", task["severity"])
	require.Equal(t, "OPEN", task["state"])
	expectedTask := backend.PRComment{ID: 13, Text: "Fix this", Severity: "BLOCKER", State: "OPEN"}
	require.Equal(t, []backend.PRComment{expectedTask}, listTasks("open"))
	require.Equal(t, []backend.PRComment{original, expectedTask}, listComments())
	reviewRecord(t, "create_pr_task", host, "success", "project", "slug", "pr_id", "body")
	reviewRecord(t, "list_pr_tasks", host, "success", "project", "slug", "pr_id", "state")
	matrixCallError(t, s, "create_pr_task", matrixWith(taskArgs, "body", "denied"), "perm.write_required", host, "new")
	require.Equal(t, []backend.PRComment{expectedTask}, listTasks("open"))
	reviewRecord(t, "create_pr_task", host, "error", "pr_id", "body")
	matrixCallError(t, s, "list_pr_tasks", matrixWith(base, "pr_id", 404), "perm.write_required", host, "404")
	reviewRecord(t, "list_pr_tasks", host, "error", "pr_id")
	resolved := matrixCallOK[map[string]any](t, s, "resolve_pr_task", matrixWith(taskArgs, "task_id", 13))
	require.Equal(t, float64(13), resolved["id"])
	require.Equal(t, "RESOLVED", resolved["state"])
	expectedTask.State = "RESOLVED"
	require.Equal(t, []backend.PRComment{expectedTask}, listTasks("resolved"))
	require.Empty(t, listTasks("open"))
	reviewRecord(t, "resolve_pr_task", host, "success", "project", "slug", "pr_id", "task_id")
	matrixCallError(t, s, "resolve_pr_task", matrixWith(taskArgs, "task_id", 99), "perm.write_required", host, "99")
	require.Equal(t, []backend.PRComment{expectedTask}, listTasks("resolved"))
	reviewRecord(t, "resolve_pr_task", host, "error", "task_id")
	reopened := matrixCallOK[map[string]any](t, s, "reopen_pr_task", matrixWith(taskArgs, "task_id", 13))
	require.Equal(t, float64(13), reopened["id"])
	require.Equal(t, "OPEN", reopened["state"])
	expectedTask.State = "OPEN"
	require.Equal(t, []backend.PRComment{expectedTask}, listTasks("open"))
	reviewRecord(t, "reopen_pr_task", host, "success", "project", "slug", "pr_id", "task_id")
	matrixCallError(t, s, "reopen_pr_task", matrixWith(taskArgs, "task_id", 99), "perm.write_required", host, "99")
	require.Equal(t, []backend.PRComment{expectedTask}, listTasks("open"))
	reviewRecord(t, "reopen_pr_task", host, "error", "task_id")
	anchored := matrixCallOK[map[string]any](t, s, "create_pr_task", matrixWith(matrixWith(taskArgs, "body", "Reply task"), "parent_comment_id", 11))
	require.Equal(t, float64(14), anchored["id"])
	require.Equal(t, "BLOCKER", anchored["severity"])
	require.Equal(t, "OPEN", anchored["state"])
	anchoredTask := backend.PRComment{ID: 14, Text: "Reply task", Severity: "BLOCKER", State: "OPEN", ParentID: 11}
	require.Equal(t, []backend.PRComment{expectedTask, anchoredTask}, listTasks("open"))
	require.Equal(t, []backend.PRComment{original, expectedTask, anchoredTask}, listComments())
	reviewRecord(t, "create_pr_task", host, "success", "project", "slug", "pr_id", "body", "parent_comment_id")
	reviewOptionalCommentInputs(t, s, host, base)
}

func reviewOptionalCommentInputs(t *testing.T, s *stdioSession, host string, base map[string]any) {
	t.Helper()
	commentArgs := matrixWith(base, "id", 42)
	listComments := func() []backend.PRComment {
		return matrixCallOK[[]backend.PRComment](t, s, "list_pr_comments", commentArgs)
	}
	before := listComments()
	require.NotEmpty(t, before)
	require.Equal(t, backend.PRComment{ID: 11, Text: "Original note"}, before[0])
	reactionsArgs := matrixWith(commentArgs, "include_reactions", true)
	if host == matrixCloudHost {
		matrixCallError(t, s, "list_pr_comments", reactionsArgs, "host.unsupported", host)
		require.Equal(t, before, listComments())
		reviewRecord(t, "list_pr_comments", host, "error", "project", "slug", "id", "include_reactions")
	} else {
		reacted := matrixCallOK[[]backend.PRComment](t, s, "list_pr_comments", reactionsArgs)
		require.Len(t, reacted, len(before))
		require.Equal(t, []backend.CommentReaction{{Emoji: "thumbs_up", Users: []backend.User{{Slug: "alice"}}}}, reacted[0].Reactions)
		require.Equal(t, before, listComments(), "reaction enrichment must not modify comments")
		reviewRecord(t, "list_pr_comments", host, "success", "project", "slug", "id", "include_reactions")
	}

	rangeArgs := matrixWith(matrixWith(matrixWith(matrixWith(commentArgs, "body", "Range note"), "inline_path", "src/sync.go"), "inline_line", 18), "inline_side", "old")
	startLine := 0
	if host == matrixCloudHost {
		startLine = 16
		rangeArgs = matrixWith(rangeArgs, "inline_start_line", startLine)
	}
	rangeComment := matrixCallOK[backend.PRComment](t, s, "add_pr_comment", rangeArgs)
	require.Equal(t, backend.PRComment{ID: 21, Text: "Range note", Inline: &backend.PRCommentInline{Path: "src/sync.go", Side: "old", Line: 18, StartLine: startLine}}, rangeComment)
	require.Equal(t, append(append([]backend.PRComment(nil), before...), rangeComment), listComments())
	flags := []string{"project", "slug", "id", "body", "inline_path", "inline_line", "inline_side"}
	if host == matrixCloudHost {
		flags = append(flags, "inline_start_line")
	}
	reviewRecord(t, "add_pr_comment", host, "success", flags...)
	if host == matrixServerHost {
		// Server rejects multi-line anchors; the MCP boundary also rejects a
		// reversed range before any backend write.
		matrixCallPlainError(t, s, "add_pr_comment", matrixWith(rangeArgs, "inline_start_line", 19), "inline_start_line must be <= inline_line")
		require.Equal(t, append(append([]backend.PRComment(nil), before...), rangeComment), listComments())
		reviewRecord(t, "add_pr_comment", host, "error", "inline_start_line")
	}
	reply := matrixCallOK[backend.PRComment](t, s, "add_pr_comment", matrixWith(matrixWith(commentArgs, "body", "Thread reply"), "parent_id", 11))
	require.Equal(t, backend.PRComment{ID: 22, Text: "Thread reply", ParentID: 11}, reply)
	require.Equal(t, append(append(append([]backend.PRComment(nil), before...), rangeComment), reply), listComments())
	reviewRecord(t, "add_pr_comment", host, "success", "project", "slug", "id", "body", "parent_id")
}
