package cloud

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/internal/paging"
)

const sshKeyPairPath = "/repositories/%s/%s/pipelines_config/ssh/key_pair"
const sshKnownHostsPath = "/repositories/%s/%s/pipelines_config/ssh/known_hosts"
const sshKnownHostPath = "/repositories/%s/%s/pipelines_config/ssh/known_hosts/%s"

// ── SSH key pair ──────────────────────────────────────────────────────────────

type cloudSSHKeyPair struct {
	PublicKey string `json:"public_key"`
	KeyType   string `json:"key_type"`
	CreatedOn string `json:"created_on"`
}

type cloudSSHKeyPairRegenBody struct {
	KeyPair struct {
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
	} `json:"key_pair"`
	Bits int `json:"bits,omitempty"`
}

func toSSHKeyPairDomain(w cloudSSHKeyPair) backend.PipelineSSHKeyPair {
	t, _ := time.Parse(time.RFC3339, w.CreatedOn)
	return backend.PipelineSSHKeyPair{
		PublicKey:    w.PublicKey,
		KeyTypeLabel: w.KeyType,
		Created:      t,
	}
}

// GetPipelineSSHKeyPair returns the SSH key pair for a repository's pipelines.
// GET /2.0/repositories/{ws}/{slug}/pipelines_config/ssh/key_pair
func (c *Client) GetPipelineSSHKeyPair(ns, slug string) (backend.PipelineSSHKeyPair, error) {
	path := fmt.Sprintf(sshKeyPairPath, url.PathEscape(ns), url.PathEscape(slug))
	var w cloudSSHKeyPair
	if err := c.getJSON(path, &w); err != nil {
		return backend.PipelineSSHKeyPair{}, err
	}
	return toSSHKeyPairDomain(w), nil
}

// RegeneratePipelineSSHKeyPair triggers generation of a new SSH key pair.
// PUT /2.0/repositories/{ws}/{slug}/pipelines_config/ssh/key_pair
// Passing bits=0 uses the Cloud default (2048).
func (c *Client) RegeneratePipelineSSHKeyPair(ns, slug string, bits int) (backend.PipelineSSHKeyPair, error) {
	path := fmt.Sprintf(sshKeyPairPath, url.PathEscape(ns), url.PathEscape(slug))
	body := cloudSSHKeyPairRegenBody{}
	if bits > 0 {
		body.Bits = bits
	}
	var w cloudSSHKeyPair
	if err := c.http.PutJSON(path, body, &w); err != nil {
		return backend.PipelineSSHKeyPair{}, err
	}
	return toSSHKeyPairDomain(w), nil
}

// ── Known hosts ───────────────────────────────────────────────────────────────

type cloudSSHPublicKey struct {
	KeyType           string `json:"key_type"`
	Key               string `json:"key"`
	MD5Fingerprint    string `json:"md5_fingerprint"`
	SHA256Fingerprint string `json:"sha256_fingerprint"`
}

type cloudKnownHost struct {
	UUID      string            `json:"uuid"`
	Hostname  string            `json:"hostname"`
	PublicKey cloudSSHPublicKey `json:"public_key"`
}

type cloudAddKnownHostBody struct {
	Hostname  string            `json:"hostname"`
	PublicKey cloudSSHPublicKey `json:"public_key"`
}

func toKnownHostDomain(w cloudKnownHost) backend.PipelineKnownHost {
	return backend.PipelineKnownHost{
		UUID:     stripBraces(w.UUID),
		Hostname: w.Hostname,
		PublicKey: backend.PipelineSSHPublicKey{
			KeyType: w.PublicKey.KeyType,
			Key:     w.PublicKey.Key,
			MD5:     w.PublicKey.MD5Fingerprint,
			SHA256:  w.PublicKey.SHA256Fingerprint,
		},
	}
}

// ListPipelineKnownHosts returns all known hosts for a repository's pipelines.
// GET /2.0/repositories/{ws}/{slug}/pipelines_config/ssh/known_hosts
func (c *Client) ListPipelineKnownHosts(ns, slug string) ([]backend.PipelineKnownHost, error) {
	path := fmt.Sprintf(sshKnownHostsPath, url.PathEscape(ns), url.PathEscape(slug))
	return paging.Collect(c.http, path, func(body []byte) ([]backend.PipelineKnownHost, error) {
		var page cloudPagedResponse[cloudKnownHost]
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, err
		}
		out := make([]backend.PipelineKnownHost, 0, len(page.Values))
		for _, w := range page.Values {
			out = append(out, toKnownHostDomain(w))
		}
		return out, nil
	}, 0)
}

// GetPipelineKnownHost returns a single known host by UUID.
// GET /2.0/repositories/{ws}/{slug}/pipelines_config/ssh/known_hosts/{uuid}
func (c *Client) GetPipelineKnownHost(ns, slug, uuid string) (backend.PipelineKnownHost, error) {
	path := fmt.Sprintf(sshKnownHostPath, url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(braceUUID(uuid)))
	var w cloudKnownHost
	if err := c.getJSON(path, &w); err != nil {
		return backend.PipelineKnownHost{}, err
	}
	return toKnownHostDomain(w), nil
}

// AddPipelineKnownHost adds a known host to a repository's pipelines config.
// POST /2.0/repositories/{ws}/{slug}/pipelines_config/ssh/known_hosts
func (c *Client) AddPipelineKnownHost(ns, slug string, in backend.PipelineKnownHostInput) (backend.PipelineKnownHost, error) {
	path := fmt.Sprintf(sshKnownHostsPath, url.PathEscape(ns), url.PathEscape(slug))
	body := cloudAddKnownHostBody{
		Hostname: in.Hostname,
		PublicKey: cloudSSHPublicKey{
			KeyType:           in.PublicKey.KeyType,
			Key:               in.PublicKey.Key,
			MD5Fingerprint:    in.PublicKey.MD5,
			SHA256Fingerprint: in.PublicKey.SHA256,
		},
	}
	var w cloudKnownHost
	if err := c.postJSON(path, body, &w); err != nil {
		return backend.PipelineKnownHost{}, err
	}
	return toKnownHostDomain(w), nil
}

// DeletePipelineKnownHost removes a known host by UUID.
// DELETE /2.0/repositories/{ws}/{slug}/pipelines_config/ssh/known_hosts/{uuid}
func (c *Client) DeletePipelineKnownHost(ns, slug, uuid string) error {
	path := fmt.Sprintf(sshKnownHostPath, url.PathEscape(ns), url.PathEscape(slug), url.PathEscape(braceUUID(uuid)))
	return c.delete(path)
}
