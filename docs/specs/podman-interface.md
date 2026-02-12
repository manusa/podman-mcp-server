# Podman Interface

## Overview

The `Podman` interface abstracts container runtime operations, allowing multiple backend implementations (CLI, REST API, etc.). A registry pattern enables implementation discovery, selection, and extensibility.

## Status

**Not yet implemented.** This spec will be implemented as part of [Podman REST API Bindings](./podman-rest-api-bindings.md).

## Requirements

- Define a `Podman` interface for container runtime operations
- Support multiple implementations via registry pattern
- Auto-detect available implementations at runtime
- Allow user override via CLI flag
- Enable future/custom implementations without modifying core code

## Architecture

### Podman Interface

The interface defines all container runtime operations:

```go
// pkg/podman/interface.go

type Podman interface {
    ContainerInspect(name string) (string, error)
    ContainerList() (string, error)
    ContainerLogs(name string) (string, error)
    ContainerRemove(name string) (string, error)
    ContainerRun(imageName string, portMappings map[int]int, envVariables []string) (string, error)
    ContainerStop(name string) (string, error)
    ImageBuild(containerFile string, imageName string) (string, error)
    ImageList() (string, error)
    ImagePull(imageName string) (string, error)
    ImagePush(imageName string) (string, error)
    ImageRemove(imageName string) (string, error)
    NetworkList() (string, error)
    VolumeList() (string, error)
}
```

### Implementation Interface

Each implementation must satisfy both `Podman` and provide metadata:

```go
// pkg/podman/registry.go

type Implementation interface {
    Podman
    Name() string        // Unique identifier (e.g., "cli", "api")
    Description() string // Human-readable description for help text
    Available() bool     // Whether this implementation can be used
    Priority() int       // Higher priority = tried first in auto-detection
}
```

### Implementation Registry

The registry follows the pattern from [kubernetes-mcp-server/pkg/toolsets/toolsets.go](https://github.com/containers/kubernetes-mcp-server/blob/main/pkg/toolsets/toolsets.go):

```go
// pkg/podman/registry.go

var implementations []Implementation

// Register adds an implementation to the registry.
// Called from init() in each implementation file.
func Register(impl Implementation)

// Implementations returns all registered implementations.
func Implementations() []Implementation

// ImplementationNames returns sorted names of all registered implementations.
func ImplementationNames() []string

// ImplementationFromString looks up an implementation by name.
func ImplementationFromString(name string) Implementation

// Clear removes all registered implementations. TESTING PURPOSES ONLY.
func Clear()
```

### Implementation Selection

```go
// pkg/podman/interface.go

// NewPodman returns a Podman implementation.
// If override is empty, auto-detects by iterating implementations by priority.
// If override is specified, returns that implementation or error if unavailable.
func NewPodman(override string) (Podman, error)
```

Selection logic:

```
                    NewPodman(override string)
                              │
                              ▼
                   ┌─────────────────────┐
                   │ Override specified? │
                   └──────────┬──────────┘
                              │
              ┌───────────────┴───────────────┐
              │ Yes                           │ No (empty)
              ▼                               ▼
     ┌─────────────────┐            ┌─────────────────────┐
     │ Lookup by name  │            │ Sort by priority    │
     │ in registry     │            │ (descending)        │
     └────────┬────────┘            └──────────┬──────────┘
              │                                │
              ▼                                ▼
     ┌─────────────────┐            ┌─────────────────────┐
     │ Found and       │            │ Return first where  │
     │ Available()?    │            │ Available() == true │
     └────────┬────────┘            └─────────────────────┘
              │
       ┌──────┴──────┐
       │ Yes    No   │
       ▼             ▼
   Return        Return error
   impl          (list available)
```

### Package Structure

```
pkg/podman/
├── interface.go          # Podman interface + NewPodman()
├── registry.go           # Implementation registry
├── podman_xxx.go         # Implementation files (cli, api, etc.)
└── socket.go             # Socket detection utilities (for api impl)
```

Each implementation file (`podman_xxx.go`) contains:
- The implementation struct
- Methods satisfying `Implementation` interface
- An `init()` function that calls `Register()`

### Registered Implementations

| Name | Description | Priority | Available When |
|------|-------------|----------|----------------|
| `api` | Podman REST API via Unix socket | 100 | Socket detected and responds to ping |
| `cli` | Podman/Docker CLI wrapper | 50 | `podman` or `docker` binary found in PATH |

Higher priority implementations are tried first during auto-detection.

## Configuration

### CLI Flag

| Flag | Description |
|------|-------------|
| `--podman-impl` | Override implementation selection (available: listed in help) |

The flag description dynamically lists available implementations using `ImplementationNames()`:

```go
// pkg/podman-mcp-server/cmd/root.go

cmd.Flags().StringVar(
    &o.PodmanImpl,
    "podman-impl",
    "",
    "Podman implementation to use (available: "+strings.Join(podman.ImplementationNames(), ", ")+"). Auto-detects if not specified.",
)
```

Example help output:

```
Flags:
  --podman-impl string   Podman implementation to use (available: api, cli). Auto-detects if not specified.
```

Example usage:

```bash
# Auto-detect (default)
podman-mcp-server

# Force CLI implementation
podman-mcp-server --podman-impl=cli

# Force API implementation (fails if socket unavailable)
podman-mcp-server --podman-impl=api
```

## Error Handling

### Auto-Detection Failure

If no implementation is available during auto-detection:
- Return error listing all implementations and their availability status
- Example: `no podman implementation available: api (socket not found), cli (binary not found)`

### Override Failure

If `--podman-impl` specifies an unavailable implementation:
- Return error with specific reason
- Example: `podman implementation "api" not available: socket not found at /run/podman/podman.sock`

### Invalid Override

If `--podman-impl` specifies an unknown implementation:
- Return error listing valid options
- Example: `invalid podman implementation "foo", valid options: api, cli`

## Testing

### Registry Isolation

Tests must isolate the registry to avoid cross-test pollution:

```go
func (s *MySuite) SetupTest() {
    podman.Clear()
    // Register only the implementations needed for this test
}
```

### Mock Implementations

Tests can register mock implementations:

```go
type mockPodman struct {
    // mock fields
}

func (m *mockPodman) Name() string        { return "mock" }
func (m *mockPodman) Description() string { return "Mock implementation for testing" }
func (m *mockPodman) Available() bool     { return true }
func (m *mockPodman) Priority() int       { return 1000 }
// ... Podman interface methods
```

## Related Code

| Component | Location |
|-----------|----------|
| Podman interface | `pkg/podman/interface.go` |
| Implementation registry | `pkg/podman/registry.go` |
| CLI flag | `pkg/podman-mcp-server/cmd/root.go` |

## Related Specs

| Spec | Relationship |
|------|--------------|
| [Podman REST API Bindings](./podman-rest-api-bindings.md) | Adds `api` implementation |

---

<details>
<summary>Design Decisions</summary>

### Why a registry pattern?

- Enables future implementations without modifying core code
- Allows users to potentially add custom implementations
- Provides clean listing of available options via CLI help
- Follows established pattern from kubernetes-mcp-server

### Why priority-based auto-detection?

- More capable implementations (API) should be preferred when available
- Simpler implementations (CLI) serve as reliable fallbacks
- Priority is explicit rather than relying on registration order

### Why `Available()` method?

- Implementations can check their prerequisites at runtime
- Avoids runtime failures by detecting unavailability early
- Enables informative error messages

### Why `Clear()` for testing only?

- Production code should never clear the registry
- Tests need isolation to avoid interference
- Naming convention makes intent clear

</details>

<details>
<summary>Future Enhancements</summary>

- Docker socket implementation (`docker` name)
- Remote Podman via SSH implementation
- Implementation-specific configuration options
- Runtime implementation switching (hot-swap)

</details>
