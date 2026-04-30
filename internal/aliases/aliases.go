// Package aliases stores user-defined command shortcuts in
// ~/.config/bitbottle/aliases.yml. Two flavors:
//   - command aliases:   `bitbottle alias set prs 'pr list --author @me'`
//   - shell aliases:     `bitbottle alias set deploy '!ssh prod ...'` (! prefix)
package aliases

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	shellquote "github.com/kballard/go-shellquote"
)

// Store is a thread-safe handle to the aliases file.
type Store struct {
	dir string
	mu  sync.Mutex
	d   map[string]string
}

func New(dir string) *Store { return &Store{dir: dir, d: map[string]string{}} }

// Load reads aliases.yml. A missing file is treated as an empty store.
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := os.ReadFile(filepath.Join(s.dir, "aliases.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			s.d = map[string]string{}
			return nil
		}
		return err
	}
	parsed := map[string]string{}
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return err
	}
	if parsed == nil {
		parsed = map[string]string{}
	}
	s.d = parsed
	return nil
}

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
	tmp, err := os.CreateTemp(s.dir, ".aliases.yml.*")
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
	return os.Rename(tmpName, filepath.Join(s.dir, "aliases.yml"))
}

// Set records expansion under name. Caller must check shadowing of built-ins
// before calling (see CheckShadow).
func (s *Store) Set(name, expansion string) error {
	if name == "" {
		return fmt.Errorf("alias name cannot be empty")
	}
	if expansion == "" {
		return fmt.Errorf("alias expansion cannot be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.d[name] = expansion
	return nil
}

func (s *Store) Get(name string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.d[name]
	return v, ok
}

func (s *Store) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.d[name]; !ok {
		return false
	}
	delete(s.d, name)
	return true
}

type Entry struct {
	Name      string
	Expansion string
}

// List returns all aliases sorted by name.
func (s *Store) List() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	names := make([]string, 0, len(s.d))
	for n := range s.d {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]Entry, 0, len(names))
	for _, n := range names {
		out = append(out, Entry{Name: n, Expansion: s.d[n]})
	}
	return out
}

// CheckShadow returns an error when name conflicts with a built-in command.
// Caller passes the list of top-level command names from the root cobra command.
func CheckShadow(name string, builtins []string) error {
	for _, b := range builtins {
		if b == name {
			return fmt.Errorf("alias %q shadows a built-in command", name)
		}
	}
	return nil
}

// Expansion describes the result of resolving an alias.
type Expansion struct {
	// Args is non-nil for command aliases. The first element is the bitbottle
	// subcommand to dispatch; remaining elements are its args. Caller's trailing
	// args (after the alias name on the command line) are appended verbatim.
	Args []string

	// Shell, when non-empty, is the literal command line to hand to $SHELL -c.
	// Args is nil in this case. Argument interpolation ($1..$9, $@) has already
	// been applied.
	Shell string
}

// Resolve looks up name in store and returns its Expansion.
//
// trailing is the slice of CLI args that followed the alias name — these are
// either appended (for command aliases) or interpolated into $1..$9 / $@ for
// shell aliases.
func Resolve(store *Store, name string, trailing []string) (*Expansion, bool, error) {
	raw, ok := store.Get(name)
	if !ok {
		return nil, false, nil
	}

	if strings.HasPrefix(raw, "!") {
		shell := strings.TrimPrefix(raw, "!")
		shell = interpolateShellArgs(shell, trailing)
		return &Expansion{Shell: shell}, true, nil
	}

	parts, err := shellquote.Split(raw)
	if err != nil {
		return nil, true, fmt.Errorf("alias %q: %w", name, err)
	}
	parts = append(parts, trailing...)
	return &Expansion{Args: parts}, true, nil
}

// interpolateShellArgs replaces $1..$9 with the corresponding trailing arg and
// $@ with all trailing args (shell-quoted). Unreferenced positions are blank.
func interpolateShellArgs(template string, args []string) string {
	out := template
	for i := 1; i <= 9; i++ {
		token := fmt.Sprintf("$%d", i)
		if !strings.Contains(out, token) {
			continue
		}
		var v string
		if i-1 < len(args) {
			v = args[i-1]
		}
		out = strings.ReplaceAll(out, token, v)
	}
	if strings.Contains(out, "$@") {
		out = strings.ReplaceAll(out, "$@", shellquote.Join(args...))
	}
	return out
}
