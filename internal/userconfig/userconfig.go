// Package userconfig stores bitbottle user preferences in ~/.config/bitbottle/config.yml.
//
// This is intentionally distinct from internal/config (which manages auth state
// in hosts.yml). User preferences cover editor/pager/browser/git_protocol/prompt
// and may be scoped per-host.
package userconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

// AllowedKeys is the set of keys writable via `bitbottle config set`. We keep
// this allowlisted so users get an immediate error on a typo rather than
// silently writing a key the CLI never reads back.
var AllowedKeys = []string{
	"editor",
	"pager",
	"browser",
	"git_protocol",
	"prompt",
}

// PerHostKeys are the subset of keys that may be overridden per-host.
var PerHostKeys = []string{
	"git_protocol",
}

// data is the on-disk YAML schema. Top-level scalars are global defaults;
// the `hosts` map holds per-host overrides keyed by hostname.
type data struct {
	Editor      string                       `yaml:"editor,omitempty"`
	Pager       string                       `yaml:"pager,omitempty"`
	Browser     string                       `yaml:"browser,omitempty"`
	GitProtocol string                       `yaml:"git_protocol,omitempty"`
	Prompt      string                       `yaml:"prompt,omitempty"`
	Hosts       map[string]map[string]string `yaml:"hosts,omitempty"`
}

// Config is a thread-safe handle to the user preferences file.
type Config struct {
	dir string
	mu  sync.Mutex
	d   data
}

// New constructs a Config rooted at dir.
func New(dir string) *Config {
	return &Config{dir: dir}
}

// Load reads config.yml. Missing file is not an error; the returned Config
// just behaves as empty.
func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	raw, err := os.ReadFile(filepath.Join(c.dir, "config.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var parsed data
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	c.d = parsed
	return nil
}

// Save writes config.yml atomically via temp file + rename.
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	bytes, err := yaml.Marshal(c.d)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(c.dir, ".config.yml.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(bytes); err != nil {
		tmp.Close()        //nolint:errcheck
		os.Remove(tmpName) //nolint:errcheck
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName) //nolint:errcheck
		return err
	}
	return os.Rename(tmpName, filepath.Join(c.dir, "config.yml"))
}

// Get returns (value, true) when the key is set. Per-host lookup falls back to
// the global value when no host-specific override exists.
func (c *Config) Get(key, host string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if host != "" {
		if hostCfg, ok := c.d.Hosts[host]; ok {
			if v, ok := hostCfg[key]; ok && v != "" {
				return v, true
			}
		}
	}
	v := globalGet(&c.d, key)
	return v, v != ""
}

// Set writes value under key (optionally scoped to host).
func (c *Config) Set(key, value, host string) error {
	if !isAllowed(key) {
		return fmt.Errorf("unknown key %q (allowed: %v)", key, AllowedKeys)
	}
	if host != "" && !isPerHost(key) {
		return fmt.Errorf("key %q cannot be set per-host", key)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if host == "" {
		globalSet(&c.d, key, value)
		return nil
	}
	if c.d.Hosts == nil {
		c.d.Hosts = map[string]map[string]string{}
	}
	if c.d.Hosts[host] == nil {
		c.d.Hosts[host] = map[string]string{}
	}
	c.d.Hosts[host][key] = value
	return nil
}

// Entry is one row returned by List.
type Entry struct {
	Key   string
	Value string
	Host  string // empty = global scope
}

// List returns every set key in deterministic order: globals first (sorted),
// then per-host (sorted by host then key).
func (c *Config) List() []Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []Entry
	for _, k := range AllowedKeys {
		if v := globalGet(&c.d, k); v != "" {
			out = append(out, Entry{Key: k, Value: v})
		}
	}
	hostnames := make([]string, 0, len(c.d.Hosts))
	for h := range c.d.Hosts {
		hostnames = append(hostnames, h)
	}
	sort.Strings(hostnames)
	for _, h := range hostnames {
		keys := make([]string, 0, len(c.d.Hosts[h]))
		for k := range c.d.Hosts[h] {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out = append(out, Entry{Key: k, Value: c.d.Hosts[h][k], Host: h})
		}
	}
	return out
}

func isAllowed(k string) bool {
	for _, a := range AllowedKeys {
		if a == k {
			return true
		}
	}
	return false
}

func isPerHost(k string) bool {
	for _, a := range PerHostKeys {
		if a == k {
			return true
		}
	}
	return false
}

func globalGet(d *data, key string) string {
	switch key {
	case "editor":
		return d.Editor
	case "pager":
		return d.Pager
	case "browser":
		return d.Browser
	case "git_protocol":
		return d.GitProtocol
	case "prompt":
		return d.Prompt
	}
	return ""
}

func globalSet(d *data, key, value string) {
	switch key {
	case "editor":
		d.Editor = value
	case "pager":
		d.Pager = value
	case "browser":
		d.Browser = value
	case "git_protocol":
		d.GitProtocol = value
	case "prompt":
		d.Prompt = value
	}
}
