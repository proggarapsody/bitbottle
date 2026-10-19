// Command qa-catalog discovers the current public CLI and MCP surfaces.
// It never invokes a Bitbucket operation.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/proggarapsody/bitbottle/internal/app"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/mcp"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

type flag struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default"`
	Required  bool   `json:"required"`
	Inherited bool   `json:"inherited,omitempty"`
}

type cliNode struct {
	ID         string   `json:"id"`
	Path       string   `json:"path"`
	Use        string   `json:"use"`
	Parameters []string `json:"parameters"`
	Short      string   `json:"short"`
	Runnable   bool     `json:"runnable"`
	Hidden     bool     `json:"hidden"`
	Flags      []flag   `json:"flags"`
	Aliases    []string `json:"aliases,omitempty"`
}

type mcpTool struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Description      string          `json:"description"`
	InputSchema      json.RawMessage `json:"input_schema"`
	Parameters       []string        `json:"parameters"`
	DeclaredBackends json.RawMessage `json:"declared_meta_backends,omitempty"`
	Raw              json.RawMessage `json:"raw"`
}

type catalog struct {
	Schema        int       `json:"schema"`
	CLI           []cliNode `json:"cli"`
	MCP           []mcpTool `json:"mcp"`
	GlobalVersion struct {
		ID       string `json:"id"`
		Observed string `json:"observed"`
	} `json:"global_version"`
}

func main() {
	switch os.Getenv("QA_CATALOG_CHILD") {
	case "version":
		os.Args = []string{"bitbottle", "--version"}
		os.Exit(app.Run())
	case "mcp":
		c := mcp.NewCmdMCP(factory.New())
		c.SetArgs([]string{"serve"})
		if err := c.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: qa-catalog")
		os.Exit(2)
	}
	cat, err := discover()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cat); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func discover() (catalog, error) {
	cat := catalog{Schema: 1, CLI: []cliNode{}, MCP: []mcpTool{}}
	r := root.NewCmdRoot(factory.New())
	r.InitDefaultHelpCmd()
	r.InitDefaultCompletionCmd()
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		path := c.CommandPath()
		flags := []flag{}
		collect := func(fs *pflag.FlagSet, inherited bool) {
			fs.VisitAll(func(f *pflag.Flag) {
				for _, known := range flags {
					if known.Name == f.Name {
						return
					}
				}
				flags = append(flags, flag{Name: f.Name, Shorthand: f.Shorthand,
					Type: f.Value.Type(), Default: f.DefValue,
					Required:  f.Annotations != nil && len(f.Annotations[cobra.BashCompOneRequiredFlag]) > 0,
					Inherited: inherited})
			})
		}
		collect(c.LocalNonPersistentFlags(), false)
		collect(c.PersistentFlags(), false)
		collect(c.InheritedFlags(), true)
		slices.SortFunc(flags, func(a, b flag) int { return strings.Compare(a.Name, b.Name) })
		cat.CLI = append(cat.CLI, cliNode{ID: "cli:" + path, Path: path,
			Use: c.Use, Parameters: strings.Fields(c.Use)[1:], Short: c.Short,
			Runnable: c.Run != nil || c.RunE != nil,
			Hidden:   c.Hidden, Flags: flags, Aliases: c.Aliases})
		for _, child := range c.Commands() {
			walk(child)
		}
	}
	walk(r)
	slices.SortFunc(cat.CLI, func(a, b cliNode) int { return strings.Compare(a.ID, b.ID) })
	version, err := childVersion()
	if err != nil {
		return cat, err
	}
	cat.GlobalVersion.ID = "cli:bitbottle --version"
	cat.GlobalVersion.Observed = version
	cat.MCP, err = discoverMCP()
	return cat, err
}

func childEnv(home, mode string) []string {
	return []string{"PATH=/usr/bin:/bin", "HOME=" + home, "TMPDIR=" + home,
		"XDG_CONFIG_HOME=" + home, "XDG_CACHE_HOME=" + home,
		"QA_CATALOG_CHILD=" + mode}
}

func childVersion() (string, error) {
	home, err := os.MkdirTemp("", "qa-catalog-version-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(home) }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, os.Args[0])
	c.Env = childEnv(home, "version")
	out, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("app.Run --version probe: %w: %s", err, out)
	}
	version := strings.TrimSpace(string(out))
	if version == "" {
		return "", fmt.Errorf("app.Run --version probe returned empty output")
	}
	return version, nil
}

func discoverMCP() ([]mcpTool, error) {
	home, err := os.MkdirTemp("", "qa-catalog-mcp-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(home) }()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, os.Args[0])
	c.Env = childEnv(home, "mcp")
	stdin, err := c.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr strings.Builder
	c.Stderr = &stderr
	if err := c.Start(); err != nil {
		return nil, err
	}
	defer func() { _ = stdin.Close(); _ = c.Process.Kill(); _ = c.Wait() }()
	protocol := client.NewClient(transport.NewIO(stdout, stdin, io.NopCloser(strings.NewReader(""))))
	if err := protocol.Start(ctx); err != nil {
		return nil, fmt.Errorf("MCP start: %w: %s", err, stderr.String())
	}
	defer func() { _ = protocol.Close() }()
	_, err = protocol.Initialize(ctx, mcplib.InitializeRequest{Params: mcplib.InitializeParams{
		ProtocolVersion: mcplib.LATEST_PROTOCOL_VERSION,
		ClientInfo:      mcplib.Implementation{Name: "qa-catalog", Version: "1"},
	}})
	if err != nil {
		return nil, fmt.Errorf("MCP initialize: %w: %s", err, stderr.String())
	}
	tools := []mcpTool{}
	seen := map[string]bool{}
	seenCursors := map[mcplib.Cursor]bool{}
	var cursor mcplib.Cursor
	for {
		res, err := protocol.ListTools(ctx, mcplib.ListToolsRequest{PaginatedRequest: mcplib.PaginatedRequest{
			Params: mcplib.PaginatedParams{Cursor: cursor},
		}})
		if err != nil {
			return nil, fmt.Errorf("MCP tools/list: %w: %s", err, stderr.String())
		}
		for _, tool := range res.Tools {
			if tool.Name == "" || seen[tool.Name] {
				return nil, fmt.Errorf("MCP duplicate or empty tool name %q", tool.Name)
			}
			seen[tool.Name] = true
			entry, err := convertMCPTool(tool)
			if err != nil {
				return nil, err
			}
			tools = append(tools, entry)
		}
		if res.NextCursor == "" {
			break
		}
		if seenCursors[res.NextCursor] {
			return nil, fmt.Errorf("MCP tools/list repeated cursor %q", res.NextCursor)
		}
		seenCursors[res.NextCursor] = true
		cursor = res.NextCursor
	}
	if len(tools) == 0 {
		return nil, fmt.Errorf("MCP server published no tools")
	}
	slices.SortFunc(tools, func(a, b mcpTool) int { return strings.Compare(a.ID, b.ID) })
	return tools, nil
}

func convertMCPTool(tool mcplib.Tool) (mcpTool, error) {
	raw, err := json.Marshal(tool)
	if err != nil {
		return mcpTool{}, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return mcpTool{}, err
	}
	var declared json.RawMessage
	if meta := fields["_meta"]; len(meta) > 0 {
		var m map[string]json.RawMessage
		if err := json.Unmarshal(meta, &m); err != nil {
			return mcpTool{}, err
		}
		declared = m["backends"]
	}
	parameters := []string{}
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(fields["inputSchema"], &schema); err != nil {
		return mcpTool{}, fmt.Errorf("MCP tool %s has invalid input schema: %w", tool.Name, err)
	}
	for name := range schema.Properties {
		parameters = append(parameters, name)
	}
	slices.Sort(parameters)
	return mcpTool{ID: "mcp:" + tool.Name, Name: tool.Name,
		Description: tool.Description, InputSchema: fields["inputSchema"], Parameters: parameters,
		DeclaredBackends: declared, Raw: raw}, nil
}
