package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// HostConfig is one entry in hosts.yml.
type HostConfig struct {
	User          string `yaml:"user"`
	OAuthToken    string `yaml:"oauth_token,omitempty"`
	GitProtocol   string `yaml:"git_protocol"`
	Version       string `yaml:"version,omitempty"`
	SkipTLSVerify bool   `yaml:"skip_tls_verify,omitempty"`
	BackendType   string `yaml:"backend_type,omitempty"`
}

// Config reads/writes hosts.yml.
type Config struct {
	dir  string
	mu   sync.Mutex
	data map[string]HostConfig
}

// New returns a new Config rooted at dir.
func New(dir string) *Config {
	return &Config{dir: dir, data: map[string]HostConfig{}}
}

// Load reads hosts.yml from disk.
func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := os.ReadFile(filepath.Join(c.dir, "hosts.yml"))
	if err != nil {
		return err
	}
	m := map[string]HostConfig{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return err
	}
	if m == nil {
		m = map[string]HostConfig{}
	}
	c.data = m
	return nil
}

// Save writes hosts.yml to disk atomically via a temp file + rename.
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := yaml.Marshal(c.data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(c.dir, ".hosts.yml.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()        //nolint:errcheck
		os.Remove(tmpName) //nolint:errcheck
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName) //nolint:errcheck
		return err
	}
	return os.Rename(tmpName, filepath.Join(c.dir, "hosts.yml"))
}

// Get returns the HostConfig for hostname, if any.
func (c *Config) Get(hostname string) (HostConfig, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	hc, ok := c.data[hostname]
	return hc, ok
}

// Set stores hc for hostname.
func (c *Config) Set(hostname string, hc HostConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[hostname] = hc
}

// Remove deletes hostname from the config.
func (c *Config) Remove(hostname string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, hostname)
}

// Hosts returns all configured hostnames.
func (c *Config) Hosts() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	hosts := make([]string, 0, len(c.data))
	for h := range c.data {
		hosts = append(hosts, h)
	}
	return hosts
}
