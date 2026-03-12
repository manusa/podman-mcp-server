package podman

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/manusa/podman-mcp-server/pkg/config"
)

// Implementation represents a Podman implementation with metadata.
// Each implementation must provide metadata for registration, discovery,
// and selection, plus a factory method to create initialized instances.
type Implementation interface {
	Name() string                             // Unique identifier (e.g., "cli", "api")
	Description() string                      // Human-readable description for help text
	Available() bool                          // Whether this implementation can be used
	Priority() int                            // Higher priority = tried first in auto-detection
	Initialize(config.Config) (Podman, error) // Creates and initializes a new instance
}

var implReg = &implRegistry{implementations: make(map[string]Implementation)}

// Register adds an implementation to the registry.
// Called from init() in each implementation file.
// Panics if an implementation is already registered with the given name.
func Register(impl Implementation) {
	implReg.register(impl)
}

// Implementations returns all registered implementations.
func Implementations() []Implementation {
	return implReg.all()
}

// ImplementationNames returns sorted names of all registered implementations.
func ImplementationNames() []string {
	return implReg.names()
}

// ImplementationFromString looks up an implementation by name.
// Returns nil if no implementation with that name is registered.
func ImplementationFromString(name string) Implementation {
	return implReg.get(name)
}

// Clear removes all registered implementations.
// TESTING PURPOSES ONLY. Production code should never clear the registry.
func Clear() {
	implReg.clear()
}

// DefaultImplementation returns the name of the default implementation.
// This is the implementation that would be selected during auto-detection
// when multiple implementations are available (highest priority wins).
// Returns empty string if no implementations are registered.
func DefaultImplementation() string {
	return implReg.defaultImpl()
}

type implRegistry struct {
	implementations map[string]Implementation
	mu              sync.RWMutex
}

func (r *implRegistry) register(impl Implementation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := impl.Name()
	if _, exists := r.implementations[name]; exists {
		panic(fmt.Sprintf("implementation already registered for name '%s'", name))
	}
	r.implementations[name] = impl
}

func (r *implRegistry) get(name string) Implementation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.implementations[name]
}

func (r *implRegistry) all() []Implementation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Implementation, 0, len(r.implementations))
	for _, impl := range r.implementations {
		result = append(result, impl)
	}
	slices.SortFunc(result, func(a, b Implementation) int {
		return strings.Compare(a.Name(), b.Name())
	})
	return result
}

func (r *implRegistry) names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.implementations))
	for name := range r.implementations {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (r *implRegistry) defaultImpl() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.implementations) == 0 {
		return ""
	}
	var best Implementation
	for _, impl := range r.implementations {
		if best == nil || impl.Priority() > best.Priority() {
			best = impl
		}
	}
	return best.Name()
}

func (r *implRegistry) clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.implementations = make(map[string]Implementation)
}

// ErrNoImplementationAvailable is returned when no implementation is available.
type ErrNoImplementationAvailable struct {
	TriedImplementations []string // Status of each implementation that was tried
}

func (e *ErrNoImplementationAvailable) Error() string {
	return "no podman implementation available: " + strings.Join(e.TriedImplementations, ", ")
}

// ErrImplementationNotAvailable is returned when a specific implementation is not available.
type ErrImplementationNotAvailable struct {
	Name   string
	Reason string
}

func (e *ErrImplementationNotAvailable) Error() string {
	return "podman implementation \"" + e.Name + "\" not available: " + e.Reason
}
