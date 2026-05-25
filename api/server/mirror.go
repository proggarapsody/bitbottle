package server

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

// serverMirrorServer is the wire representation of a Smart Mirror server.
type serverMirrorServer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"baseUrl"`
	Enabled bool   `json:"enabled"`
}

// serverMirroredRepo is the wire representation of a mirrored repository.
type serverMirroredRepo struct {
	Slug     string `json:"slug"`
	MirrorId string `json:"mirrorId"`
	Status   string `json:"status"`
	LastSync int64  `json:"lastSync"` // epoch ms, may be 0
}

func toMirrorServerDomain(s serverMirrorServer) backend.MirrorServer {
	return backend.MirrorServer{
		ID:      s.ID,
		Name:    s.Name,
		BaseURL: s.BaseURL,
		Enabled: s.Enabled,
	}
}

func toMirroredRepo(s serverMirroredRepo) backend.MirroredRepo {
	var lastSync time.Time
	if s.LastSync > 0 {
		lastSync = time.UnixMilli(s.LastSync).UTC()
	}
	return backend.MirroredRepo{
		Slug:       s.Slug,
		MirrorID:   s.MirrorId,
		LastSyncAt: lastSync,
		Status:     s.Status,
	}
}

// ListMirrorServers returns all Smart Mirror servers.
// GET /rest/mirroring/latest/mirrorServers
func (c *Client) ListMirrorServers(limit int) ([]backend.MirrorServer, error) {
	return paging.Collect(c.mirrorHTTP, "/mirrorServers", func(body []byte) ([]backend.MirrorServer, error) {
		var page PagedResponse[serverMirrorServer]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.MirrorServer, 0, len(page.Values))
		for _, s := range page.Values {
			out = append(out, toMirrorServerDomain(s))
		}
		return out, nil
	}, limit)
}

// GetMirrorServer returns a single Smart Mirror server by ID.
// GET /rest/mirroring/latest/mirrorServers/{id}
func (c *Client) GetMirrorServer(id string) (backend.MirrorServer, error) {
	var s serverMirrorServer
	path := fmt.Sprintf("/mirrorServers/%s", url.PathEscape(id))
	if err := c.mirrorHTTP.GetJSON(path, &s); err != nil {
		return backend.MirrorServer{}, err
	}
	return toMirrorServerDomain(s), nil
}

// ListMirroredRepos returns all repos mirrored by the given mirror server.
// GET /rest/mirroring/latest/mirrorServers/{mirrorID}/repos
func (c *Client) ListMirroredRepos(mirrorID string, limit int) ([]backend.MirroredRepo, error) {
	path := fmt.Sprintf("/mirrorServers/%s/repos", url.PathEscape(mirrorID))
	return paging.Collect(c.mirrorHTTP, path, func(body []byte) ([]backend.MirroredRepo, error) {
		var page PagedResponse[serverMirroredRepo]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.MirroredRepo, 0, len(page.Values))
		for _, s := range page.Values {
			out = append(out, toMirroredRepo(s))
		}
		return out, nil
	}, limit)
}
