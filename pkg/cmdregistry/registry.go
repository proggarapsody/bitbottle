// Package cmdregistry provides a self-registration mechanism for Cobra
// subcommands. New command packages call Register (typically from init()) so
// they appear in the root command without editing root.go.
//
// The package exposes both an instance-based Registry (useful for testing) and
// a package-level global (used by production init() registrations).
package cmdregistry

import (
	"sort"
	"sync"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Builder is a function that constructs a *cobra.Command given a factory.
type Builder func(*factory.Factory) *cobra.Command

// Registry holds a set of command builders and can materialise them in
// deterministic order.
type Registry struct {
	mu       sync.Mutex
	builders []Builder
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register appends a builder to the registry. Safe to call from init().
func (r *Registry) Register(b Builder) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.builders = append(r.builders, b)
}

// All applies every registered builder to f and returns the resulting commands
// sorted alphabetically by Use. The sort is deterministic regardless of init()
// order, which varies across build systems and test binaries.
func (r *Registry) All(f *factory.Factory) []*cobra.Command {
	r.mu.Lock()
	snapshot := make([]Builder, len(r.builders))
	copy(snapshot, r.builders)
	r.mu.Unlock()

	cmds := make([]*cobra.Command, 0, len(snapshot))
	for _, b := range snapshot {
		cmds = append(cmds, b(f))
	}
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Use < cmds[j].Use
	})
	return cmds
}

// global is the package-level registry used by production init() calls.
var global = NewRegistry()

// Register adds a builder to the package-level global registry.
// Call this from init() in any command package.
func Register(b Builder) {
	global.Register(b)
}

// All returns all commands registered in the package-level global registry,
// sorted alphabetically by Use. Called once from root.go after the fixed
// AddCommand block.
func All(f *factory.Factory) []*cobra.Command {
	return global.All(f)
}
