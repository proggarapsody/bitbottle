// Package profiles stores named credential profiles in
// ~/.config/bitbottle/profiles.yml. Each profile is a set of Bitbucket
// credentials (hostname, token, user, etc.) that can be switched as a unit,
// similar to how kubectl manages contexts.
package profiles

import (
	"os"
	"path/filepath"
	"sort"
	"sync"

	"gopkg.in/yaml.v3"
)

// Profile is a named set of Bitbucket credentials.
type Profile struct {
	Hostname      string `yaml:"hostname"`
	Token         string `yaml:"token"`
	User          string `yaml:"user,omitempty"`
	AuthUser      string `yaml:"auth_user,omitempty"`
	SkipTLSVerify bool   `yaml:"skip_tls_verify,omitempty"`
	BackendType   string `yaml:"backend_type,omitempty"` // "server" | "cloud" | "" (auto)
	GitProtocol   string `yaml:"git_protocol,omitempty"` // "https" | "ssh"
}

// NamedProfile is a Profile annotated with its store key, used for listing.
type NamedProfile struct {
	Name string
	Profile
}

// Store is a thread-safe handle to profiles.yml.
type Store struct {
	dir string
	mu  sync.Mutex
	d   map[string]Profile // key = profile name
}

// New returns a new Store rooted at dir.
func New(dir string) *Store {
	return &Store{dir: dir, d: map[string]Profile{}}
}

// Load reads profiles.yml. A missing file is treated as an empty store.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := os.ReadFile(filepath.Join(s.dir, "profiles.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			s.d = map[string]Profile{}
			return nil
		}
		return err
	}
	parsed := map[string]Profile{}
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	if parsed == nil {
		parsed = map[string]Profile{}
	}
	s.d = parsed
	return nil
}

// Save writes profiles.yml atomically via a temp file + rename.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	bytes, err := yaml.Marshal(s.d)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.dir, ".profiles.yml.*")
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
	return os.Rename(tmpName, filepath.Join(s.dir, "profiles.yml"))
}

// Get returns the Profile for name, if any.
func (s *Store) Get(name string) (Profile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.d[name]
	return p, ok
}

// Set stores p under name, overwriting any existing profile with that name.
func (s *Store) Set(name string, p Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.d[name] = p
}

// Delete removes the profile with the given name. Returns true if it existed.
func (s *Store) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.d[name]; !ok {
		return false
	}
	delete(s.d, name)
	return true
}

// All returns all profiles as a sorted slice of NamedProfile.
func (s *Store) All() []NamedProfile {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.d))
	for n := range s.d {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]NamedProfile, 0, len(names))
	for _, n := range names {
		out = append(out, NamedProfile{Name: n, Profile: s.d[n]})
	}
	return out
}
