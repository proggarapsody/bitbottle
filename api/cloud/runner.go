package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const runnersPath = "/workspaces/%s/pipelines-config/runners"
const runnerPath = "/workspaces/%s/pipelines-config/runners/%s"

// wireRunner is the Bitbucket Cloud JSON shape for a pipeline runner.
type wireRunner struct {
	UUID      string          `json:"uuid"`
	Name      string          `json:"name"`
	State     wireRunnerState `json:"state"`
	Platform  wirePlatform    `json:"platform"`
	Labels    []wireLabel     `json:"labels"`
	CreatedOn string          `json:"created_on"`
}

type wireRunnerState struct {
	Status string `json:"status"`
}

type wirePlatform struct {
	OperatingSystem string `json:"operating_system"`
	Architecture    string `json:"architecture"`
}

type wireLabel struct {
	Name string `json:"name"`
}

// wireRunnerCreate is the request body for POST .../pipelines-config/runners.
type wireRunnerCreate struct {
	Name     string       `json:"name"`
	Labels   []wireLabel  `json:"labels"`
	Platform wirePlatform `json:"platform"`
}

// normalizeArch maps the API's architecture token to the canonical bitbottle form.
// The API uses "X86_64"; bitbottle uses "AMD64".
func normalizeArch(arch string) string {
	if strings.EqualFold(arch, "X86_64") {
		return "AMD64"
	}
	return strings.ToUpper(arch)
}

// apiArch maps the canonical bitbottle arch to the API's token.
func apiArch(arch string) string {
	if strings.EqualFold(arch, "AMD64") {
		return "X86_64"
	}
	return strings.ToUpper(arch)
}

func toRunnerDomain(w wireRunner) backend.Runner {
	labels := make([]string, 0, len(w.Labels))
	for _, l := range w.Labels {
		labels = append(labels, l.Name)
	}
	return backend.Runner{
		UUID:  stripBraces(w.UUID),
		Name:  w.Name,
		State: w.State.Status,
		Platform: backend.RunnerPlatform{
			Operating: strings.ToUpper(w.Platform.OperatingSystem),
			Arch:      normalizeArch(w.Platform.Architecture),
		},
		Labels:    labels,
		CreatedOn: w.CreatedOn,
	}
}

// ListRunners returns all pipeline self-hosted runners for a workspace.
func (c *Client) ListRunners(workspace string) ([]backend.Runner, error) {
	path := fmt.Sprintf(runnersPath, workspace)
	return paging.Collect(
		c.http,
		path,
		func(body []byte) ([]backend.Runner, error) {
			var page struct {
				Values []wireRunner `json:"values"`
			}
			if err := json.Unmarshal(body, &page); err != nil {
				return nil, err
			}
			out := make([]backend.Runner, 0, len(page.Values))
			for _, w := range page.Values {
				out = append(out, toRunnerDomain(w))
			}
			return out, nil
		},
		0, // unbounded
	)
}

// CreateRunner registers a new self-hosted runner in a workspace.
func (c *Client) CreateRunner(workspace string, in backend.CreateRunnerInput) (backend.Runner, error) {
	labels := make([]wireLabel, 0, len(in.Labels))
	for _, l := range in.Labels {
		labels = append(labels, wireLabel{Name: l})
	}
	body := wireRunnerCreate{
		Name:   in.Name,
		Labels: labels,
		Platform: wirePlatform{
			OperatingSystem: strings.ToUpper(in.Platform.Operating),
			Architecture:    apiArch(in.Platform.Arch),
		},
	}
	var w wireRunner
	path := fmt.Sprintf(runnersPath, workspace)
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.Runner{}, err
	}
	return toRunnerDomain(w), nil
}

// DeleteRunner removes a pipeline self-hosted runner from a workspace.
func (c *Client) DeleteRunner(workspace, runnerUUID string) error {
	path := fmt.Sprintf(runnerPath, workspace, braceUUID(runnerUUID))
	return c.delete(path)
}
