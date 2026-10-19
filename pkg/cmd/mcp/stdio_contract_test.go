package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	cmdmcp "github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const stdioChildEnv = "BITBOTTLE_TEST_MCP_STDIO_CHILD"
const matrixChildEnv = "BITBOTTLE_TEST_MCP_MATRIX_CHILD"

// TestMCPStdioChild re-executes this test binary so the command owns real
// process stdin/stdout, while only the Bitbucket backend is replaced.
func TestMCPStdioChild(t *testing.T) {
	if os.Getenv(stdioChildEnv) != "1" {
		return
	}
	if os.Getenv(matrixChildEnv) == "1" {
		runMatrixChild(t)
		return
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{
		ConfigDir:     os.Getenv("BITBOTTLE_TEST_MCP_CONFIG_DIR"),
		InitialConfig: "git.example.com:\n  oauth_token: test-token\n",
	})
	factorytest.UseBackend(f, &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(project, slug string, id int) (backend.PullRequest, error) {
			if project != "PROJ" || slug != "repo" {
				return backend.PullRequest{}, fmt.Errorf("unexpected repository %s/%s", project, slug)
			}
			switch id {
			case 42:
				return backend.PullRequest{ID: 42, Title: "Known fixture PR", State: "OPEN"}, nil
			case 404:
				return backend.PullRequest{}, &backend.DomainError{
					Kind: backend.ErrNotFound, Code: backend.CodePRNotFound,
					Host: "git.example.com", Resource: "pull-request", ID: "404",
					Message: "pull request #404 not found",
				}
			default:
				return backend.PullRequest{}, fmt.Errorf("unexpected pull request id %d", id)
			}
		},
	})
	cmd := cmdmcp.NewCmdMCP(f)
	cmd.SetArgs([]string{"serve"})
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// The Go test runner would print PASS on stdout after this function returns.
	os.Exit(0)
}

func TestMCPStdioContract(t *testing.T) {
	s := startStdioSession(t, false)
	ctx := s.ctx
	protocolClient := s.client

	initialized := s.initialized
	require.Equal(t, "bitbottle", initialized.ServerInfo.Name)
	require.Contains(t, mcplib.ValidProtocolVersions, initialized.ProtocolVersion)
	require.NotNil(t, initialized.Capabilities.Tools)

	listed, err := protocolClient.ListTools(ctx, mcplib.ListToolsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, listed.Tools)
	t.Logf("MCP stdio published %d tools", len(listed.Tools))
	byName := make(map[string]mcplib.Tool, len(listed.Tools))
	for _, tool := range listed.Tools {
		require.NotEmpty(t, tool.Name)
		require.NotEmpty(t, tool.Description, "tool %q has no description", tool.Name)
		require.NotContains(t, byName, tool.Name, "duplicate published tool")
		byName[tool.Name] = tool
	}
	assertToolFields(t, byName, "get_pr", map[string]string{
		"project": "string", "slug": "string", "id": "number",
	}, "project", "slug", "id")
	assertToolFields(t, byName, "add_pr_comment", map[string]string{
		"project": "string", "slug": "string", "id": "number", "body": "string",
	}, "project", "slug", "id", "body")

	call := func(name string, args map[string]any) (*mcplib.CallToolResult, error) {
		return s.call(name, args)
	}
	validArgs := map[string]any{"hostname": "git.example.com", "project": "PROJ", "slug": "repo", "id": 42}
	result, err := call("get_pr", validArgs)
	require.NoError(t, err)
	require.False(t, result.IsError)
	var pr backend.PullRequest
	require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &pr))
	require.Equal(t, backend.PullRequest{ID: 42, Title: "Known fixture PR", State: "OPEN"}, pr)

	notFoundArgs := map[string]any{"hostname": "git.example.com", "project": "PROJ", "slug": "repo", "id": 404}
	result, err = call("get_pr", notFoundArgs)
	require.NoError(t, err)
	require.True(t, result.IsError)
	var domainFailure struct {
		Code     string   `json:"code"`
		Host     string   `json:"host"`
		Resource string   `json:"resource"`
		ID       string   `json:"id"`
		Hints    []string `json:"hints"`
	}
	require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &domainFailure))
	require.Equal(t, "pr.not_found", domainFailure.Code)
	require.Equal(t, "git.example.com", domainFailure.Host)
	require.Equal(t, "pull-request", domainFailure.Resource)
	require.Equal(t, "404", domainFailure.ID)
	require.NotEmpty(t, domainFailure.Hints)

	for _, tc := range []struct {
		name  string
		args  map[string]any
		code  string
		field string
	}{
		{"missing id", map[string]any{"project": "PROJ", "slug": "repo"}, "arg.missing", "id"},
		{"invalid id type", map[string]any{"project": "PROJ", "slug": "repo", "id": "forty-two"}, "arg.invalid_type", "id"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := call("get_pr", tc.args)
			require.NoError(t, err)
			require.True(t, result.IsError)
			var failure struct {
				Code  string `json:"code"`
				Field string `json:"field"`
			}
			require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &failure))
			require.Equal(t, tc.code, failure.Code)
			require.Equal(t, tc.field, failure.Field)
		})
	}

	result, err = call("no_such_bitbottle_tool", nil)
	require.Nil(t, result)
	require.True(t, errors.Is(err, mcplib.ErrInvalidParams), "unknown tool must be a JSON-RPC invalid-params error: %v", err)
	require.ErrorContains(t, err, "not found")
	for _, tc := range []struct {
		name string
		args map[string]any
	}{
		{"update_issue", map[string]any{"project": "PROJ", "slug": "repo", "id": "forty-two", "title": "Updated"}},
		{"reopen_issue", map[string]any{"project": "PROJ", "slug": "repo", "id": "forty-two"}},
		{"assign_issue", map[string]any{"project": "PROJ", "slug": "repo", "id": "forty-two", "assignee": "alice"}},
	} {
		t.Run(tc.name+" rejects string id", func(t *testing.T) {
			result, err := call(tc.name, tc.args)
			require.NoError(t, err)
			require.True(t, result.IsError)
			var failure struct {
				Code  string `json:"code"`
				Field string `json:"field"`
			}
			require.NoError(t, json.Unmarshal([]byte(resultText(t, result)), &failure))
			require.Equal(t, "arg.invalid_type", failure.Code)
			require.Equal(t, "id", failure.Field)
		})
	}

	s.close(t, len(listed.Tools))
}

type stdioSession struct {
	ctx         context.Context
	client      *client.Client
	initialized *mcplib.InitializeResult
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	clientInput *io.PipeReader
	relayOutput *io.PipeWriter
	relayDone   chan error
	rawStdout   *bytes.Buffer
	stderr      *bytes.Buffer
	waited      bool
}

func startStdioSession(t *testing.T, matrix bool) *stdioSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	configDir := t.TempDir()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMCPStdioChild$")
	cmd.Env = []string{
		"PATH=/usr/bin:/bin",
		"HOME=" + configDir,
		"TMPDIR=" + configDir,
		stdioChildEnv + "=1",
		"BITBOTTLE_TEST_MCP_CONFIG_DIR=" + configDir,
	}
	if matrix {
		cmd.Env = append(cmd.Env, matrixChildEnv+"=1")
	}
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

	// Relay every byte from real child stdout into the SDK and a separate
	// capture. The SDK silently skips malformed lines, so the capture is
	// independently checked after the process exits.
	clientInput, relayOutput := io.Pipe()
	s.clientInput, s.relayOutput = clientInput, relayOutput
	go func() {
		_, copyErr := io.Copy(io.MultiWriter(relayOutput, s.rawStdout), stdout)
		_ = relayOutput.CloseWithError(copyErr)
		s.relayDone <- copyErr
	}()
	protocolClient := client.NewClient(transport.NewIO(clientInput, stdin, io.NopCloser(strings.NewReader(""))))
	s.client = protocolClient
	t.Cleanup(func() {
		_ = protocolClient.Close()
		_ = clientInput.Close()
		_ = relayOutput.Close()
	})
	require.NoError(t, protocolClient.Start(ctx))

	initialized, err := protocolClient.Initialize(ctx, mcplib.InitializeRequest{
		Params: mcplib.InitializeParams{
			ProtocolVersion: mcplib.LATEST_PROTOCOL_VERSION,
			ClientInfo:      mcplib.Implementation{Name: "bitbottle-stdio-contract-test", Version: "1"},
		},
	})
	require.NoError(t, err)
	s.initialized = initialized
	return s
}

func (s *stdioSession) call(name string, args map[string]any) (*mcplib.CallToolResult, error) {
	return s.client.CallTool(s.ctx, mcplib.CallToolRequest{
		Params: mcplib.CallToolParams{Name: name, Arguments: args},
	})
}

func (s *stdioSession) close(t *testing.T, wantTools int) {
	t.Helper()
	require.NoError(t, s.stdin.Close())
	select {
	case err := <-s.relayDone:
		require.NoError(t, err)
	case <-s.ctx.Done():
		t.Fatalf("MCP child stdout did not close: %v; stderr: %s", s.ctx.Err(), s.stderr.String())
	}
	require.NoError(t, s.cmd.Wait(), "MCP child stderr: %s", s.stderr.String())
	s.waited = true
	assertJSONRPCStdout(t, s.rawStdout.Bytes(), wantTools)
}

func assertToolFields(t *testing.T, tools map[string]mcplib.Tool, name string, fields map[string]string, required ...string) {
	t.Helper()
	tool, ok := tools[name]
	require.True(t, ok, "missing published tool %q", name)
	require.Equal(t, "object", tool.InputSchema.Type)
	for field, wantType := range fields {
		property, ok := tool.InputSchema.Properties[field].(map[string]any)
		require.True(t, ok, "%s.%s missing property", name, field)
		require.Equal(t, wantType, property["type"], "%s.%s type", name, field)
		require.NotEmpty(t, property["description"], "%s.%s description", name, field)
		require.Contains(t, tool.InputSchema.Required, field, "%s.%s required", name, field)
	}
	require.ElementsMatch(t, required, tool.InputSchema.Required, "%s required fields", name)
}

func resultText(t *testing.T, result *mcplib.CallToolResult) string {
	t.Helper()
	require.Len(t, result.Content, 1)
	content, ok := result.Content[0].(mcplib.TextContent)
	require.True(t, ok)
	return content.Text
}

func assertJSONRPCStdout(t *testing.T, output []byte, wantTools int) {
	t.Helper()
	require.NotEmpty(t, output)
	require.Equal(t, byte('\n'), output[len(output)-1], "stdout must contain complete JSON-RPC lines")
	responseCount := 0
	toolsWithCompiledSchemas := 0
	lines := bytes.Split(output, []byte{'\n'})
	for lineNumber, line := range lines {
		if lineNumber == len(lines)-1 {
			continue
		}
		require.NotEmpty(t, bytes.TrimSpace(line), "stdout line %d is blank", lineNumber+1)
		var message map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(line, &message), "stdout line %d: %q", lineNumber+1, line)
		var version string
		require.NoError(t, json.Unmarshal(message["jsonrpc"], &version), "stdout line %d", lineNumber+1)
		require.Equal(t, "2.0", version, "stdout line %d", lineNumber+1)
		if id, hasID := message["id"]; hasID {
			require.NotEqual(t, "null", string(id), "stdout line %d", lineNumber+1)
			_, hasResult := message["result"]
			_, hasError := message["error"]
			require.NotEqual(t, hasResult, hasError, "stdout line %d must have one result or error", lineNumber+1)
			if hasResult {
				var result map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(message["result"], &result), "stdout line %d", lineNumber+1)
				if rawTools, hasTools := result["tools"]; hasTools {
					var tools []struct {
						Name        string          `json:"name"`
						InputSchema json.RawMessage `json:"inputSchema"`
					}
					require.NoError(t, json.Unmarshal(rawTools, &tools), "stdout line %d", lineNumber+1)
					for _, tool := range tools {
						var schemaDocument any
						require.NoError(t, json.Unmarshal(tool.InputSchema, &schemaDocument), "tool %q schema", tool.Name)
						compiler := jsonschema.NewCompiler()
						uri := fmt.Sprintf("https://bitbottle.test/tool-schema/%d", toolsWithCompiledSchemas)
						require.NoError(t, compiler.AddResource(uri, schemaDocument), "tool %q schema", tool.Name)
						_, err := compiler.Compile(uri)
						require.NoError(t, err, "tool %q schema", tool.Name)
						toolsWithCompiledSchemas++
					}
				}
			}
			responseCount++
		} else {
			require.Contains(t, message, "method", "stdout line %d must be a notification", lineNumber+1)
		}
	}
	require.GreaterOrEqual(t, responseCount, 7)
	require.Equal(t, wantTools, toolsWithCompiledSchemas, "every listed tool must publish a compiling input schema")
}
