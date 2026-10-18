package acceptancecloud

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	mcpclient "github.com/mark3labs/mcp-go/client"
	mcptransport "github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type repoData struct {
	UUID        string `json:"uuid"`
	Description string `json:"description"`
	MainBranch  struct {
		Name string `json:"name"`
	} `json:"mainbranch"`
}
type prData struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Source      struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"source"`
	Destination struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	} `json:"destination"`
}
type commentsData struct {
	Values []struct {
		Content struct {
			Raw string `json:"raw"`
		} `json:"content"`
	} `json:"values"`
}

func run(ctx context.Context, cfg Config) (report Report, retErr error) {
	start := clock(cfg)
	report = Report{Version: 1, Backend: "cloud", Scenario: ScenarioID, ScenarioIDs: []string{ScenarioID}, RunCount: 1, BackendVersion: "Bitbucket Cloud REST API v2; provider build unavailable", Status: "infrastructure_failure", IdentitySeparation: "unverified; bootstrap and actor credentials may belong to the same account", Checks: map[string]Check{}}
	defer func() {
		report.Duration = clock(cfg).Sub(start).String()
		if retErr != nil {
			report.Error = sanitize(retErr.Error(), cfg.BootstrapToken, cfg.ActorToken)
		}
		report.Checks["report_persistence"] = Check{Status: "pass"}
		countReport(&report)
		if writeErr := writeReport(cfg.ReportDir, report); writeErr != nil {
			report.Checks["report_persistence"] = Check{Status: "fail", Detail: sanitize(writeErr.Error(), cfg.BootstrapToken, cfg.ActorToken)}
			if report.Status == "pass" {
				report.Status = "infrastructure_failure"
			}
			persistenceErr := fmt.Errorf("acceptance report persistence failed: %w", writeErr)
			if retErr == nil {
				retErr = persistenceErr
			} else {
				retErr = errors.Join(retErr, persistenceErr)
			}
			report.Error = sanitize(retErr.Error(), cfg.BootstrapToken, cfg.ActorToken)
			countReport(&report)
		}
	}()
	if err := validate(cfg); err != nil {
		return report, err
	}
	report.CommitSHA = firstNonEmpty(os.Getenv("BB_ACCEPTANCE_SHA"), os.Getenv("GITHUB_SHA"), currentSHA())
	api, err := newAPI(cfg.BaseURL, cfg.BootstrapToken)
	if err != nil {
		return report, err
	}
	actor, err := newAPI(cfg.BaseURL, cfg.ActorToken)
	if err != nil {
		return report, err
	}
	runid, err := randomHex(8)
	if err != nil {
		return report, err
	}
	stamp := clock(cfg).UTC().Format("20060102t150405")
	slug := fmt.Sprintf("bbqa-%s-%s", stamp, runid)
	marker := "bitbottle-qa-run:" + runid
	manifestPath := filepath.Join(cfg.ReportDir, "manifest.json")
	manifest := Manifest{Version: 1, RunID: runid, CreatedAt: clock(cfg).UTC(), Resource: Resource{Workspace: cfg.Workspace, Slug: slug, Marker: marker}}
	if err = os.MkdirAll(cfg.ReportDir, 0700); err != nil {
		return report, fmt.Errorf("report directory: %w", err)
	}
	if err = checkAbsent(ctx, api, cfg.Workspace, slug); err != nil {
		return report, err
	}
	if err = manifest.Write(manifestPath); err != nil {
		return report, fmt.Errorf("persist cleanup manifest: %w", err)
	}
	report.Manifest = manifestPath
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		cs := clock(cfg)
		e := cleanup(cleanupCtx, manifest, cfg.BootstrapToken, cfg.BaseURL)
		if e != nil {
			report.Cleanup = Check{Status: "fail", Duration: clock(cfg).Sub(cs).String(), Detail: sanitize(e.Error(), cfg.BootstrapToken, cfg.ActorToken)}
			if retErr == nil {
				retErr = fmt.Errorf("cleanup failed: %w", e)
			}
			report.Status = "cleanup_failure"
		} else {
			report.Cleanup = Check{Status: "pass", Duration: clock(cfg).Sub(cs).String()}
		}
		countReport(&report)
	}()

	step := func(name string, fn func() error) error {
		st := clock(cfg)
		e := fn()
		status := "pass"
		detail := ""
		if e != nil {
			status = "fail"
			detail = sanitize(e.Error(), cfg.BootstrapToken, cfg.ActorToken)
			if strings.HasPrefix(name, "cli_") || strings.HasPrefix(name, "mcp_") {
				report.Status = "product_failure"
			}
		}
		report.Checks[name] = Check{Status: status, Duration: clock(cfg).Sub(st).String(), Detail: detail}
		return e
	}
	if err = step("bootstrap", func() error {
		payload := map[string]any{"scm": "git", "is_private": true, "description": marker, "project": map[string]string{"key": cfg.Project}}
		var r repoData
		_, e := api.json(ctx, "POST", repoPath(cfg.Workspace, slug), payload, &r)
		if e != nil {
			return e
		}
		manifest.Resource.UUID = r.UUID
		return manifest.Write(manifestPath)
	}); err != nil {
		return report, err
	}
	if err = step("seed", func() error { return seed(ctx, api, manifest.Resource) }); err != nil {
		return report, err
	}
	prid := 0
	if err = step("cli_pr_create", func() error {
		env, cleanupHome, e := actorEnv(cfg)
		if e != nil {
			return e
		}
		defer cleanupHome()
		prTitle := "Acceptance " + manifest.RunID
		cmd := exec.CommandContext(ctx, cfg.Binary, "pr", "create", "--repo", cfg.Workspace+"/"+slug, "--title", prTitle, "--body", marker, "--base", "main", "--head", sourceBranch(manifest.RunID), "--json", "id")
		cmd.Env = env
		out, e := cmd.CombinedOutput()
		if e != nil {
			return fmt.Errorf("CLI PR create failed (exit): %s", sanitize(strings.TrimSpace(string(out)), cfg.BootstrapToken, cfg.ActorToken))
		}
		var parsed struct {
			ID int `json:"id"`
		}
		if e = json.Unmarshal(out, &parsed); e != nil || parsed.ID < 1 {
			return errors.New("CLI PR create returned no JSON id")
		}
		prid = parsed.ID
		manifest.Resource.PRIDs = []int{prid}
		return manifest.Write(manifestPath)
	}); err != nil {
		return report, err
	}
	if err = step("cli_repo_view", func() error {
		return runCLI(ctx, cfg, manifest.Resource, "repo", "view", cfg.Workspace+"/"+slug, "--json", "slug,description")
	}); err != nil {
		return report, err
	}
	if err = step("cli_effect", func() error {
		return verifyPR(ctx, actor, manifest.Resource, prid, "Acceptance "+manifest.RunID, marker)
	}); err != nil {
		return report, err
	}
	if err = step("mcp_initialize_list_and_comment", func() error {
		// mcp-go owns framing, initialization, tool listing, and calls.
		env, cleanupHome, e := actorEnv(cfg)
		if e != nil {
			return e
		}
		defer cleanupHome()
		cli, e := mcpclient.NewStdioMCPClientWithOptions(cfg.Binary, env, []string{"mcp", "serve"}, mcpCommandEnv(env))
		if e != nil {
			return e
		}
		defer func() { _ = cli.Close() }()
		_, e = cli.Initialize(ctx, mcp.InitializeRequest{Params: mcp.InitializeParams{ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION, ClientInfo: mcp.Implementation{Name: "bitbottle-acceptance", Version: "1"}}})
		if e != nil {
			return e
		}
		listed, e := cli.ListTools(ctx, mcp.ListToolsRequest{})
		if e != nil {
			return e
		}
		found := false
		getPR := false
		for _, tool := range listed.Tools {
			if tool.Name == "add_pr_comment" {
				found = true
			}
			if tool.Name == "get_pr" {
				getPR = true
			}
		}
		if !found {
			return errors.New("MCP tool add_pr_comment is not registered")
		}
		if !getPR {
			return errors.New("MCP tool get_pr is not registered")
		}
		getArgs := map[string]any{"project": cfg.Workspace, "slug": manifest.Resource.Slug, "id": prid}
		prResult, e := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "get_pr", Arguments: getArgs}})
		if e != nil {
			return e
		}
		if prResult.IsError {
			return errors.New("MCP get_pr returned an error")
		}
		if !mcpTextContains(prResult, "Acceptance "+manifest.RunID) {
			return errors.New("MCP get_pr result did not contain the expected PR title")
		}
		args := map[string]any{"project": cfg.Workspace, "slug": manifest.Resource.Slug, "id": prid, "body": "MCP acceptance comment " + manifest.RunID}
		res, e := cli.CallTool(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "add_pr_comment", Arguments: args}})
		if e != nil {
			return e
		}
		if res.IsError {
			return errors.New("MCP add_pr_comment returned an error")
		}
		return verifyComment(ctx, actor, manifest.Resource, prid, "MCP acceptance comment "+manifest.RunID)
	}); err != nil {
		return report, err
	}
	report.Status = "pass"
	return report, nil
}

func mcpCommandEnv(env []string) mcptransport.StdioOption {
	childEnv := append([]string(nil), env...)
	return mcptransport.WithCommandFunc(func(ctx context.Context, command string, _ []string, args []string) (*exec.Cmd, error) {
		cmd := exec.CommandContext(ctx, command, args...)
		cmd.Env = append([]string(nil), childEnv...)
		return cmd, nil
	})
}

func validate(c Config) error {
	if c.BootstrapToken == "" || c.ActorToken == "" || c.Email == "" || c.Workspace != "bitbottle-qa-lab" || c.Project != "BBQA" || c.ReportDir == "" || c.Binary == "" {
		return errors.New("cloud acceptance requires bootstrap token, distinct actor token, email, allowlisted workspace/project, report directory, and binary")
	}
	if c.BootstrapToken == c.ActorToken {
		return errors.New("BB_CLOUD_BOOTSTRAP_TOKEN and BB_TOKEN must be different credentials")
	}
	if _, e := exec.LookPath(c.Binary); e != nil {
		return errors.New("BITBOTTLE_BIN must point to an executable")
	}
	if err := os.MkdirAll(c.ReportDir, 0700); err != nil {
		return errors.New("BB_ACCEPTANCE_REPORT_DIR is not writable")
	}
	f, e := os.CreateTemp(c.ReportDir, ".write-check-")
	if e != nil {
		return errors.New("BB_ACCEPTANCE_REPORT_DIR is not writable")
	}
	_ = f.Close()
	_ = os.Remove(f.Name())
	return nil
}

func checkAbsent(ctx context.Context, c *apiClient, w, s string) error {
	var r repoData
	_, err := c.json(ctx, "GET", repoPath(w, s), nil, &r)
	if err == nil {
		return errors.New("generated Cloud repository slug already exists")
	}
	if strings.Contains(err.Error(), "HTTP 404") {
		return nil
	}
	return err
}
func seed(ctx context.Context, c *apiClient, r Resource) error {
	root := repoPath(r.Workspace, r.Slug)
	if err := c.upload(ctx, root+"/src", "main", "Bitbottle QA acceptance seed", "README.md", "Synthetic acceptance fixture\n"); err != nil {
		return err
	}
	var b struct {
		Hash string `json:"hash"`
	}
	if _, err := c.json(ctx, "GET", root+"/commit/main", nil, &b); err != nil {
		return err
	}
	if b.Hash == "" {
		return errors.New("seed commit hash missing")
	}
	branch := sourceBranch(strings.TrimPrefix(r.Marker, "bitbottle-qa-run:"))
	payload := map[string]any{"name": branch, "target": map[string]string{"hash": b.Hash}}
	if _, err := c.json(ctx, "POST", root+"/refs/branches", payload, nil); err != nil {
		return err
	}
	return c.upload(ctx, root+"/src", branch, "Bitbottle QA acceptance change", "acceptance.txt", "Synthetic source branch content\n")
}
func verifyPR(ctx context.Context, c *apiClient, r Resource, id int, title, marker string) error {
	var p prData
	_, e := c.json(ctx, "GET", repoPath(r.Workspace, r.Slug)+fmt.Sprintf("/pullrequests/%d", id), nil, &p)
	if e != nil {
		return e
	}
	if p.ID != id || p.Title != title || p.Description != marker || p.Source.Branch.Name != sourceBranch(strings.TrimPrefix(marker, "bitbottle-qa-run:")) || p.Destination.Branch.Name != "main" || p.State != "OPEN" {
		return errors.New("independent REST PR verification failed")
	}
	return nil
}

func mcpTextContains(result *mcp.CallToolResult, part string) bool {
	for _, item := range result.Content {
		if text, ok := item.(mcp.TextContent); ok && strings.Contains(text.Text, part) {
			return true
		}
	}
	return false
}
func verifyComment(ctx context.Context, c *apiClient, r Resource, id int, body string) error {
	var p commentsData
	_, e := c.json(ctx, "GET", repoPath(r.Workspace, r.Slug)+fmt.Sprintf("/pullrequests/%d/comments?pagelen=100", id), nil, &p)
	if e != nil {
		return e
	}
	for _, v := range p.Values {
		if v.Content.Raw == body {
			return nil
		}
	}
	return errors.New("independent REST comment verification failed")
}

func actorEnv(cfg Config) ([]string, func(), error) {
	root, err := os.MkdirTemp("", "bbqa-isolated-")
	if err != nil {
		return nil, func() {}, err
	}
	configDir := filepath.Join(root, "bitbottle")
	if err = os.MkdirAll(configDir, 0700); err != nil {
		_ = os.RemoveAll(root)
		return nil, func() {}, err
	}
	// Identity and host selection are metadata only. The actor token is passed
	// exclusively through BB_TOKEN and never written to this isolated config.
	metadata := fmt.Sprintf("bitbucket.org:\n  user: %q\n  auth_user: %q\n  backend_type: cloud\n  git_protocol: https\n", cfg.Email, cfg.Email)
	if err = os.WriteFile(filepath.Join(configDir, "hosts.yml"), []byte(metadata), 0600); err != nil {
		_ = os.RemoveAll(root)
		return nil, func() {}, err
	}
	env := filteredEnv(append(os.Environ(), "HOME="+root, "XDG_CONFIG_HOME="+root, "BB_HOST=bitbucket.org", "BB_TOKEN="+cfg.ActorToken))
	return env, func() { _ = os.RemoveAll(root) }, nil
}

func runCLI(ctx context.Context, cfg Config, r Resource, args ...string) error {
	env, cleanupHome, err := actorEnv(cfg)
	if err != nil {
		return err
	}
	defer cleanupHome()
	cmd := exec.CommandContext(ctx, cfg.Binary, args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CLI command failed: %s", sanitize(strings.TrimSpace(string(out)), cfg.BootstrapToken, cfg.ActorToken))
	}
	if len(args) >= 4 && args[0] == "repo" && args[1] == "view" {
		var result struct {
			Slug        string `json:"slug"`
			Description string `json:"description"`
		}
		if err = json.Unmarshal(out, &result); err != nil {
			return errors.New("CLI repo view returned invalid JSON")
		}
		if result.Slug != r.Slug || result.Description != r.Marker {
			return errors.New("CLI repo view data verification failed")
		}
	}
	return nil
}
func repoPath(w, s string) string    { return "/repositories/" + esc(w) + "/" + esc(s) }
func sourceBranch(run string) string { return "acceptance/" + run }
func clock(c Config) time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}
func currentSHA() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	b, e := exec.CommandContext(ctx, "git", "rev-parse", "HEAD").Output()
	if e != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
func countReport(r *Report) {
	r.PassCount = 0
	r.FailCount = 0
	r.SkipCount = 0
	if r.Status == "pass" {
		r.PassCount = 1
	} else {
		r.FailCount = 1
	}
}
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return hex.EncodeToString(b), nil
}
func filteredEnv(env []string) []string {
	allowed := map[string]bool{"PATH": true, "HOME": true, "TMPDIR": true, "TMP": true, "TEMP": true, "LANG": true, "LC_ALL": true, "LC_CTYPE": true, "SYSTEMROOT": true, "WINDIR": true, "XDG_CONFIG_HOME": true, "BB_HOST": true, "BB_TOKEN": true}
	values := map[string]string{}
	for _, entry := range env {
		k, v, ok := strings.Cut(entry, "=")
		if ok && (allowed[k] || strings.HasPrefix(k, "LC_")) {
			values[k] = v
		}
	}
	out := make([]string, 0, len(values))
	for k, v := range values {
		out = append(out, k+"="+v)
	}
	return out
}
func sanitize(s string, secrets ...string) string {
	for _, x := range secrets {
		if x != "" {
			s = strings.ReplaceAll(s, x, "[redacted]")
		}
	}
	if len(s) > 500 {
		return s[:500]
	}
	return s
}
func writeReport(dir string, r Report) error {
	if dir == "" {
		return nil
	}
	b, e := json.MarshalIndent(r, "", "  ")
	if e != nil {
		return e
	}
	tmp, e := os.CreateTemp(dir, ".report-*.json")
	if e != nil {
		return e
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if e = tmp.Chmod(0600); e != nil {
		_ = tmp.Close()
		return e
	}
	if _, e = tmp.Write(append(b, '\n')); e != nil {
		_ = tmp.Close()
		return e
	}
	if e = tmp.Sync(); e != nil {
		_ = tmp.Close()
		return e
	}
	if e = tmp.Close(); e != nil {
		return e
	}
	if e = os.Rename(name, filepath.Join(dir, "report.json")); e != nil {
		return e
	}
	return nil
}
