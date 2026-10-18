package acceptancecloud

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ScenarioID = "cloud-pr-cli-mcp-lifecycle"

type Config struct {
	BootstrapToken, ActorToken, Email, Workspace, Project string
	ReportDir, Binary                                     string
	BaseURL                                               string // Set only by hermetic tests; production callers leave empty.
	Now                                                   func() time.Time
}

type Resource struct {
	Workspace string `json:"workspace"`
	Slug      string `json:"slug"`
	UUID      string `json:"uuid,omitempty"`
	Marker    string `json:"marker"`
	PRIDs     []int  `json:"pr_ids,omitempty"`
}

type Manifest struct {
	Version   int       `json:"version"`
	RunID     string    `json:"run_id"`
	CreatedAt time.Time `json:"created_at"`
	Resource  Resource  `json:"resource"`
}

type Check struct {
	Status   string `json:"status"`
	Duration string `json:"duration,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

type Report struct {
	Version            int              `json:"version"`
	CommitSHA          string           `json:"commit_sha"`
	Backend            string           `json:"backend"`
	Scenario           string           `json:"scenario"`
	Status             string           `json:"status"`
	ScenarioIDs        []string         `json:"scenario_ids"`
	RunCount           int              `json:"run_count"`
	PassCount          int              `json:"pass_count"`
	FailCount          int              `json:"fail_count"`
	SkipCount          int              `json:"skip_count"`
	BackendVersion     string           `json:"backend_version"`
	IdentitySeparation string           `json:"identity_separation"`
	Checks             map[string]Check `json:"checks"`
	Manifest           string           `json:"manifest"`
	Cleanup            Check            `json:"cleanup"`
	Duration           string           `json:"duration"`
	Error              string           `json:"error,omitempty"`
}

func (m Manifest) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}

func ReadManifest(path string) (Manifest, error) {
	var m Manifest
	b, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	if err = ValidateManifest(m); err != nil {
		return m, err
	}
	return m, nil
}

func ValidateManifest(m Manifest) error {
	if m.Version != 1 || m.RunID == "" || m.Resource.Workspace == "" || m.Resource.Slug == "" || m.Resource.Marker == "" {
		return errors.New("invalid Cloud acceptance manifest")
	}
	if len(m.RunID) != 16 || m.Resource.Marker != "bitbottle-qa-run:"+m.RunID {
		return errors.New("manifest run ownership marker is invalid")
	}
	for _, r := range m.RunID {
		hexChar := r >= 'a' && r <= 'f' || r >= '0' && r <= '9'
		if !hexChar {
			return errors.New("manifest run ID is invalid")
		}
	}
	if m.Resource.Workspace != "bitbottle-qa-lab" || len(m.Resource.Slug) < 6 || len(m.Resource.Slug) > 63 || m.Resource.Slug[:5] != "bbqa-" {
		return errors.New("manifest target is outside the Cloud acceptance allowlist")
	}
	for _, r := range m.Resource.Slug {
		slugChar := r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-'
		if !slugChar {
			return errors.New("manifest repository slug is invalid")
		}
	}
	if !strings.HasSuffix(m.Resource.Slug, "-"+m.RunID) {
		return errors.New("manifest repository slug does not match its run ID")
	}
	return nil
}

// Run creates an isolated Cloud repository, exercises the public CLI and MCP
// client, independently verifies the effects, then always attempts cleanup.
func Run(ctx context.Context, cfg Config) (Report, error) { return run(ctx, cfg) }

// Cleanup removes only a repository described by a validated ownership manifest.
func Cleanup(ctx context.Context, manifestPath, bootstrapToken, baseURL string) error {
	m, err := ReadManifest(manifestPath)
	if err != nil {
		return err
	}
	return cleanup(ctx, m, bootstrapToken, baseURL)
}
