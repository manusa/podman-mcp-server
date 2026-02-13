package podman

import (
	"sort"
	"strings"
	"sync"
)

// Implementation represents a Podman implementation with metadata.
// Each implementation must satisfy both the Podman interface and provide
// metadata for registration, discovery, and selection.
type Implementation interface {
	Podman
	Name() string        // Unique identifier (e.g., "cli", "api")
	Description() string // Human-readable description for help text
	Available() bool     // Whether this implementation can be used
	Priority() int       // Higher priority = tried first in auto-detection
}

var (
	implementations []Implementation
	mu              sync.RWMutex
)

// Register adds an implementation to the registry.
// Called from init() in each implementation file.
// Implementations are stored in registration order; sorting by priority
// happens at selection time.
func Register(impl Implementation) {
	mu.Lock()
	defer mu.Unlock()
	implementations = append(implementations, impl)
}

// Implementations returns all registered implementations.
func Implementations() []Implementation {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Implementation, len(implementations))
	copy(result, implementations)
	return result
}

// ImplementationNames returns sorted names of all registered implementations.
func ImplementationNames() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, len(implementations))
	for i, impl := range implementations {
		names[i] = impl.Name()
	}
	sort.Strings(names)
	return names
}

// ImplementationFromString looks up an implementation by name.
// Returns nil if no implementation with that name is registered.
func ImplementationFromString(name string) Implementation {
	mu.RLock()
	defer mu.RUnlock()
	for _, impl := range implementations {
		if impl.Name() == name {
			return impl
		}
	}
	return nil
}

// Clear removes all registered implementations.
// TESTING PURPOSES ONLY. Production code should never clear the registry.
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	implementations = nil
}

// ErrNoImplementationAvailable is returned when no implementation is available.
type ErrNoImplementationAvailable struct {
	Details []string // Status of each implementation
}

func (e *ErrNoImplementationAvailable) Error() string {
	return "no podman implementation available: " + strings.Join(e.Details, ", ")
}

// ErrImplementationNotAvailable is returned when a specific implementation is not available.
type ErrImplementationNotAvailable struct {
	Name   string
	Reason string
}

func (e *ErrImplementationNotAvailable) Error() string {
	return "podman implementation \"" + e.Name + "\" not available: " + e.Reason
}
