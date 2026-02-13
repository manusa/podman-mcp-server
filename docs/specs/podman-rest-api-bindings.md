# Podman REST API Bindings

## Overview

An implementation of the [`Podman` interface](./podman-interface.md) that uses the official Podman Go bindings (`pkg/bindings`) to communicate with the Podman REST API via Unix socket, instead of shelling out to the CLI.

## Status

**Not yet implemented.** See [Implementation Plan](#implementation-plan) for phased rollout.

## Requirements

- Implement the `Podman` interface using `github.com/containers/podman/v5/pkg/bindings`
- Register as `api` implementation with higher priority than CLI
- Auto-detect Podman socket availability
- Support `CONTAINER_HOST` environment variable for socket path
- Maintain full compatibility with existing MCP tools (no behavior changes)
- Build with `remote` tag to minimize binary size increase

## Architecture

### Implementation Registration

The API implementation registers itself following the [Podman Interface](./podman-interface.md) registry pattern:

```go
// pkg/podman/podman_api.go

func init() {
    Register(&podmanApi{})
}

type podmanApi struct {
    ctx context.Context  // Context with connection info
}

func (p *podmanApi) Name() string        { return "api" }
func (p *podmanApi) Description() string { return "Podman REST API via Unix socket" }
func (p *podmanApi) Priority() int       { return 100 }  // Higher than CLI (50)

func (p *podmanApi) Available() bool {
    socketPath, err := DetectSocket()
    if err != nil {
        return false
    }
    return PingSocket(socketPath) == nil
}
```

### Socket Detection

Default socket locations to check (in order):

| Source | Path |
|--------|------|
| `CONTAINER_HOST` env var | Value from environment (e.g., `unix:///run/podman/podman.sock`) |
| Rootful default | `/run/podman/podman.sock` |
| Rootless default | `$XDG_RUNTIME_DIR/podman/podman.sock` |
| Rootless fallback | `/run/user/<UID>/podman/podman.sock` |

```go
// pkg/podman/socket.go

// DetectSocket returns the first available socket path.
func DetectSocket() (string, error)

// PingSocket verifies the socket responds to API requests.
func PingSocket(socketPath string) error
```

### Connection Lifecycle

The implementation establishes a connection once when first used and reuses it:

```go
type podmanApi struct {
    ctx      context.Context  // Context with connection (from bindings.NewConnection)
    initOnce sync.Once
    initErr  error
}

func (p *podmanApi) ensureConnection() error {
    p.initOnce.Do(func() {
        socketPath, err := DetectSocket()
        if err != nil {
            p.initErr = err
            return
        }
        p.ctx, p.initErr = bindings.NewConnection(context.Background(), socketPath)
    })
    return p.initErr
}
```

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CONTAINER_HOST` | Podman socket URI (e.g., `unix:///run/podman/podman.sock`) |

### CLI Flag

See [Podman Interface: CLI Flag](./podman-interface.md#cli-flag) for the `--podman-impl` flag.

```bash
# Force API implementation
podman-mcp-server --podman-impl=api
```

## API Mapping

Each `Podman` interface method maps to a `pkg/bindings` function:

| Interface Method | Bindings Package | Bindings Function |
|------------------|------------------|-------------------|
| `ContainerInspect(name)` | `containers` | `Inspect(ctx, name, opts)` |
| `ContainerList()` | `containers` | `List(ctx, opts)` |
| `ContainerLogs(name)` | `containers` | `Logs(ctx, name, opts)` |
| `ContainerRemove(name)` | `containers` | `Remove(ctx, name, opts)` |
| `ContainerRun(...)` | `containers` | `CreateWithSpec()` + `Start()` |
| `ContainerStop(name)` | `containers` | `Stop(ctx, name, opts)` |
| `ImageBuild(...)` | `images` | `Build(ctx, files, opts)` |
| `ImageList()` | `images` | `List(ctx, opts)` |
| `ImagePull(name)` | `images` | `Pull(ctx, name, opts)` |
| `ImagePush(name)` | `images` | `Push(ctx, name, opts)` |
| `ImageRemove(name)` | `images` | `Remove(ctx, names, opts)` |
| `NetworkList()` | `network` | `List(ctx, opts)` |
| `VolumeList()` | `volumes` | `List(ctx, opts)` |

## Testing Strategy

See [Podman Interface: Testing](./podman-interface.md#testing) for registry isolation patterns.

### Integration Tests

All MCP tool tests in `pkg/mcp` run against both CLI and API implementations using parameterized test suites:

```go
func TestContainerToolsWithAPI(t *testing.T) {
    suite.Run(t, &ContainerToolsSuite{
        McpSuite: test.McpSuite{PodmanImpl: "api"},
    })
}
```

### Mock Server

The existing mock Podman server in `internal/test/mock_server.go` serves both CLI and API implementations since both ultimately call the Podman REST API.

### Unit Tests

Socket detection has dedicated unit tests in `pkg/podman/socket_test.go`.

## Build Configuration

Use the `remote` build tag to exclude local Podman operation code and minimize binary size:

```bash
go build -tags "remote exclude_graphdriver_btrfs btrfs_noversion exclude_graphdriver_devicemapper containers_image_openpgp"
```

## Error Handling

### Connection Errors

If socket connection fails:
- `Available()` returns `false`
- Auto-detection falls back to CLI implementation
- Explicit `--podman-impl=api` returns error with socket path and reason

### Runtime Errors

If an API call fails after successful connection:
- Return the error to the caller (same as CLI behavior)
- Do NOT fall back to CLI (could cause inconsistent state)

## Dependencies

New dependency required:

```go
require github.com/containers/podman/v5 v5.x.x
```

## Related Code

| Component | Location |
|-----------|----------|
| API implementation | `pkg/podman/podman_api.go` (new) |
| Socket detection | `pkg/podman/socket.go` (new) |
| Test infrastructure | `internal/test/mcp.go` |

## Related Specs

| Spec | Relationship |
|------|--------------|
| [Podman Interface](./podman-interface.md) | Defines the interface and registry this implementation uses |

---

## Implementation Plan

> **IMPORTANT - Ephemeral Section**
>
> This section tracks implementation progress and will be **deleted once implementation is complete**.
>
> **Maintainer Instructions:**
> - Update "Current State" after completing each phase
> - Keep both this spec and [podman-interface.md](./podman-interface.md) synchronized with implementation changes
> - Once Phase 4 is complete and verified, remove this entire "Implementation Plan" section

### Current State

**Phase: 2 (Socket Detection and API Implementation - Read-Only) - Complete**

Socket detection and read-only API implementation are now in place:
- `pkg/podman/socket.go` with `DetectSocket()` and `PingSocket()` functions
- Unit tests in `pkg/podman/socket_test.go`
- `github.com/containers/podman/v5` dependency added
- `pkg/podman/podman_api.go` with `podmanApi` struct implementing `Implementation`
- Read-only methods implemented: `ContainerList()`, `ContainerInspect()`, `ContainerLogs()`, `ImageList()`, `NetworkList()`, `VolumeList()`
- API registered via `init()` with priority 100 (higher than CLI's 50)
- Write operations return "not yet implemented" error (Phase 3)
- Test infrastructure updated to default to CLI for mock server compatibility
- All tests pass

### Phase 0: Test Infrastructure Enhancement ✓

**Goal:** Enable testing multiple implementations through the same test suites.

**Spec:** This phase primarily affects test infrastructure, prerequisites for both specs.

**Status: COMPLETE**

1. ✓ Updated `McpSuite` in `internal/test/mcp.go`:
   - Added `PodmanImpl string` field
   - Added mechanism to override implementation at test time
   - Mock server works for both CLI and API patterns

2. ✓ Refactored infrastructure for parameterized tests:
   - Added `AvailableImplementations()` helper function
   - Added `DefaultImplementation()` helper function
   - Existing CLI tests continue to pass

**Deliverable:** Test infrastructure ready for multi-implementation testing.

### Phase 1: Implementation Registry ✓

**Goal:** Add registry pattern per [podman-interface.md](./podman-interface.md).

**Spec:** [Podman Interface](./podman-interface.md)

**Status: COMPLETE**

1. ✓ Created `pkg/podman/registry.go` with:
   - `Implementation` interface (`Name()`, `Description()`, `Available()`, `Priority()`)
   - `Register()`, `Implementations()`, `ImplementationNames()`, `ImplementationFromString()`
   - `Clear()` for testing
   - Error types: `ErrNoImplementationAvailable`, `ErrImplementationNotAvailable`

2. ✓ Refactored existing CLI implementation to use registry:
   - Added `Name()`, `Description()`, `Available()`, `Priority()` methods
   - Registered via `init()` function with priority 50

3. ✓ Updated `NewPodman()` in `interface.go`:
   - Uses registry for implementation selection
   - Auto-detects by priority when no override specified
   - Returns specific error types for different failure modes

4. ✓ Added `--podman-impl` CLI flag with dynamic help text listing implementations

**Deliverable:** Registry pattern works, CLI implementation registered, flag available.

### Phase 2: Socket Detection and API Implementation (Read-Only) ✓

**Goal:** Add socket detection and implement read-only API operations.

**Spec:** This spec (Podman REST API Bindings)

**Status: COMPLETE**

1. ✓ Created `pkg/podman/socket.go` with:
   - `DetectSocket() (string, error)`
   - `PingSocket(socketPath string) error`
   - Unit tests in `pkg/podman/socket_test.go`

2. ✓ Added `github.com/containers/podman/v5` dependency

3. ✓ Created `pkg/podman/podman_api.go` with:
   - `podmanApi` struct implementing `Implementation`
   - Read-only methods: `ContainerList()`, `ContainerInspect()`, `ContainerLogs()`, `ImageList()`, `NetworkList()`, `VolumeList()`
   - Registered via `init()` with priority 100

4. ✓ Updated test infrastructure to default to CLI for mock server compatibility

**Deliverable:** Read-only operations work via API when socket available.

### Phase 3: API Implementation (Write Operations)

**Goal:** Implement mutating operations.

**Spec:** This spec (Podman REST API Bindings)

1. Add remaining methods to `podmanApi`:
   - `ContainerRun(...)`, `ContainerStop()`, `ContainerRemove()`
   - `ImagePull()`, `ImagePush()`, `ImageRemove()`, `ImageBuild()`

2. Handle short-name image resolution (docker.io prefix fallback)

3. Add/update integration tests for write operations

**Deliverable:** Full API implementation complete.

### Phase 4: Documentation and Finalization

**Goal:** Complete documentation and clean up.

**Specs:** Both specs

1. Update documentation:
   - README.md with `--podman-impl` flag
   - AGENTS.md Podman Interface section

2. Update spec statuses from "Not yet implemented" to implemented

3. Update AGENTS.md spec table statuses from "Planned" to "Implemented"

4. **Delete this Implementation Plan section from both specs**

5. Update build system if needed for `remote` tag

**Deliverable:** Feature complete, documented, specs finalized.

---

<details>
<summary>Design Decisions</summary>

### Why use `pkg/bindings` instead of `libpod`?

- `pkg/bindings` is officially supported and stable
- `libpod` is explicitly marked as unstable for external use
- `pkg/bindings` has minimal native dependencies with proper build tags
- Socket-based approach enables future remote Podman support

### Why not embed the Podman API server?

Embedding would require importing `libpod` (unstable) and all its native dependencies. The socket approach keeps the dependency footprint small.

### Why lazy connection initialization?

- `Available()` can be called without establishing a full connection
- Connection established only when implementation is actually used
- Avoids unnecessary socket connections during auto-detection

### Why not fall back to CLI on operation failure?

Mixing implementations mid-session could cause inconsistent state. If the API connection was established successfully, stick with it. If it fails, the error should propagate so users can investigate.

</details>

<details>
<summary>Future Enhancements</summary>

- Remote Podman support via SSH (`ssh://user@host/run/podman/podman.sock`)
- Connection pooling for high-throughput scenarios
- Health check endpoint for socket status
- Configurable connection timeout

</details>

<details>
<summary>Research Summary</summary>

### Why Podman Can't Work Like Helm

Helm's `pkg/action` is a pure Go library with no daemon requirement. Podman's architecture requires system-level access (cgroups, namespaces, OCI runtime) that cannot be safely embedded.

### Official Guidance

From Podman maintainers ([GitHub Discussion #9076](https://github.com/containers/podman/discussions/9076)):

> "The only supported 'bindings' are the RESTful bindings. There are not a 'specific set of bindings' available for consumption with golang--only socket based."

> "It is highly unlikely rootless Podman will ever support a non-socket-based API."

### libpod Stability Warning

From [pkg.go.dev/libpod](https://pkg.go.dev/github.com/containers/podman/v5/libpod):

> "The libpod library is not stable and we do not support use cases outside of this repository. The API can change at any time even with patch releases."

### References

- [Podman Go Bindings README](https://github.com/containers/podman/blob/main/pkg/bindings/README.md)
- [pkg/bindings Documentation](https://pkg.go.dev/github.com/containers/podman/v5/pkg/bindings)
- [Podman REST API Reference](https://docs.podman.io/en/latest/_static/api.html)

</details>
