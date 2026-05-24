package server

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// Wire types for the BBS pull-request settings endpoint.
// These are not in the oapi-codegen gen/ file because they are not yet
// part of the generated OpenAPI surface for this project.

type restPRSettingsStrategy struct {
	ID string `json:"id"`
}

type restPRSettingsMergeConfig struct {
	DefaultStrategy restPRSettingsStrategy   `json:"defaultStrategy"`
	Strategies      []restPRSettingsStrategy `json:"strategies"`
}

type restPRSettings struct {
	RequiredApprovers        int                       `json:"requiredApprovers"`
	RequiredAllApprovers     bool                      `json:"requiredAllApprovers"`
	RequiredAllTasksComplete bool                      `json:"requiredAllTasksComplete"`
	RequiredSuccessfulBuilds int                       `json:"requiredSuccessfulBuilds"`
	MergeConfig              restPRSettingsMergeConfig `json:"mergeConfig"`
}

func toPRSettingsDomain(w restPRSettings) backend.RepoPRSettings {
	strategies := make([]string, 0, len(w.MergeConfig.Strategies))
	for _, s := range w.MergeConfig.Strategies {
		strategies = append(strategies, s.ID)
	}
	return backend.RepoPRSettings{
		RequiredApprovers:        w.RequiredApprovers,
		RequiredAllApprovers:     w.RequiredAllApprovers,
		RequiredAllTasksComplete: w.RequiredAllTasksComplete,
		RequiredSuccessfulBuilds: w.RequiredSuccessfulBuilds,
		MergeStrategy:            w.MergeConfig.DefaultStrategy.ID,
		AllowedStrategies:        strategies,
	}
}

func prSettingsPath(ns, slug string) string {
	return fmt.Sprintf("/projects/%s/repos/%s/settings/pull-requests",
		url.PathEscape(ns), url.PathEscape(slug))
}

// GetRepoPRSettings returns the current pull-request gate settings for a
// repository on Bitbucket Server / Data Center.
func (c *Client) GetRepoPRSettings(ns, slug string) (backend.RepoPRSettings, error) {
	var w restPRSettings
	if err := c.getJSON(prSettingsPath(ns, slug), &w); err != nil {
		return backend.RepoPRSettings{}, err
	}
	return toPRSettingsDomain(w), nil
}

// UpdateRepoPRSettings updates the pull-request gate settings for a repository
// on Bitbucket Server / Data Center. Nil pointer fields in in are left unchanged
// by first fetching the current settings and merging the non-nil fields.
func (c *Client) UpdateRepoPRSettings(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
	path := prSettingsPath(ns, slug)

	// Fetch current settings so we can merge partial updates.
	var current restPRSettings
	if err := c.getJSON(path, &current); err != nil {
		return backend.RepoPRSettings{}, err
	}

	if in.RequiredApprovers != nil {
		current.RequiredApprovers = *in.RequiredApprovers
	}
	if in.RequiredAllApprovers != nil {
		current.RequiredAllApprovers = *in.RequiredAllApprovers
	}
	if in.RequiredAllTasksComplete != nil {
		current.RequiredAllTasksComplete = *in.RequiredAllTasksComplete
	}
	if in.RequiredSuccessfulBuilds != nil {
		current.RequiredSuccessfulBuilds = *in.RequiredSuccessfulBuilds
	}
	if in.MergeStrategy != nil {
		current.MergeConfig.DefaultStrategy.ID = *in.MergeStrategy
	}
	if in.AllowedStrategies != nil {
		strats := make([]restPRSettingsStrategy, 0, len(*in.AllowedStrategies))
		for _, s := range *in.AllowedStrategies {
			strats = append(strats, restPRSettingsStrategy{ID: s})
		}
		current.MergeConfig.Strategies = strats
	}

	var result restPRSettings
	if err := c.postJSON(path, current, &result); err != nil {
		return backend.RepoPRSettings{}, err
	}
	return toPRSettingsDomain(result), nil
}
