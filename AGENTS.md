# Project Agents.md for Podman MCP Server

This Agents.md file provides comprehensive guidance for AI assistants and coding agents (like Claude, Gemini, Cursor, and others) to work with this codebase.

This repository contains the podman-mcp-server project,
a Go-based Model Context Protocol (MCP) server that provides container management capabilities using Podman or Docker.
This MCP server enables AI assistants (like Claude, Gemini, Cursor, and others) to interact with container runtimes using the Model Context Protocol (MCP).

## Project Structure and Repository layout

- Go package layout follows the standard Go conventions:
  - `cmd/podman-mcp-server/` – main application entry point.
  - `pkg/` – libraries grouped by domain.
    - `api/` - SDK-agnostic types for tool definitions (`ServerTool`, `ToolHandlerFunc`, `ToolHandlerParams`).
    - `config/` - Server configuration (`Config` struct, defaults, and override merging).
    - `mcp/` - Model Context Protocol (MCP) server implementation using the official Go SDK, with tool definitions for containers, images, networks, and volumes.
    - `podman/` - Podman/Docker abstraction layer with interface definition, implementation registry, CLI and REST API implementations.
    - `podman-mcp-server/cmd/` - CLI command definition using Cobra framework.
    - `version/` - Version information management.
  - `internal/test/` – shared test utilities (McpSuite, mock Podman API server).
- `.github/` – GitHub-related configuration (Actions workflows, Dependabot).
- `build/` – modular Makefile includes for packaging targets.
  - `node.mk` – NPM packaging targets (npm-copy-binaries, npm-copy-project-files, npm-publish).
  - `python.mk` – Python/PyPI packaging targets (python-publish).
- `npm/` – Node packages that wrap the compiled binaries for distribution through npmjs.com.
- `python/` – Python package providing a script that downloads the correct platform binary from the GitHub releases page and runs it for distribution through pypi.org.
- `Makefile` – core build tasks; includes `build/*.mk` for packaging targets.
- `server.json` – MCP Registry manifest file for publishing to the official registry.

## Feature development

Implement new functionality in the Go sources under `cmd/` and `pkg/`.
The JavaScript (`npm/`) and Python (`python/`) directories only wrap the compiled binary for distribution (npm and PyPI).
Most changes will not require touching them unless the version or packaging needs to be updated.

### Adding new MCP tools

Tools are currently organized by resource type in `pkg/mcp/`:

- `podman_container.go` - Container tools (inspect, list, logs, remove, run, stop)
- `podman_image.go` - Image tools (build, list, pull, push, remove)
- `podman_network.go` - Network tools (list)
- `podman_volume.go` - Volume tools (list)

When adding a new tool:
1. Identify the appropriate resource file (or create a new one).
2. Define the tool using `api.ServerTool` with the SDK-agnostic types from `pkg/api/`.
3. Implement the handler function using the signature `func(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error)`.
4. Add the tool to the `initXTools()` function in the resource file (e.g., `initContainerTools()`).
5. If creating a new resource type, register the `initXTools()` function in `mcp.go` within `NewServer()`.
6. Add tests for the new tool.

Example tool definition:

```go
api.ServerTool{
    Tool: api.Tool{
        Name:        "container_list",
        Description: "List all containers",
        Annotations: api.ToolAnnotations{
            Title:        "List Containers",
            ReadOnlyHint: ptr(true),
        },
        InputSchema: api.InputSchema{
            Type:       "object",
            Properties: map[string]api.Property{
                "all": {Type: "boolean", Description: "Show all containers"},
            },
        },
    },
    Handler: func(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
        all := params.GetString("all", "false") == "true"
        result, err := params.Podman.ContainerList(ctx, all)
        return api.NewToolCallResult(result, err), nil
    },
}
```

### SDK Architecture

The project uses a layered architecture that decouples tool definitions from the MCP SDK:

- **`pkg/api/`** - SDK-agnostic types (`ServerTool`, `Tool`, `ToolHandlerParams`, `ToolCallResult`)
- **`pkg/mcp/gosdk.go`** - Conversion layer between internal types and the official Go SDK
- **`pkg/mcp/mcp.go`** - Server wiring that registers tools with the go-sdk

This design allows tool definitions to be written without depending on any specific SDK, making it easier to support multiple transports or SDK versions.

### Podman Interface

The `pkg/podman/interface.go` file defines the `Podman` interface that abstracts container runtime operations.
A registry pattern in `pkg/podman/registry.go` enables multiple implementations with auto-detection.

Available implementations:
- **`cli`** (`pkg/podman/podman_cli.go`) - Uses Podman/Docker CLI commands (priority: 50). Available when `podman` or `docker` binary is in PATH.
- **`api`** (`pkg/podman/podman_api.go`) - Uses Podman REST API via Unix socket with `pkg/bindings` (priority: 100). Available when a Podman socket is detected and responds to ping. Can be excluded from builds with the `exclude_podman_api` build tag.

Socket detection (`pkg/podman/socket.go`) checks these locations in order: `CONTAINER_HOST` env var, `/run/podman/podman.sock`, `$XDG_RUNTIME_DIR/podman/podman.sock`, `/run/user/<UID>/podman/podman.sock`.

When adding new container operations:
1. Add the method signature to the `Podman` interface in `interface.go`.
2. Implement the method in both `podman_cli.go` and `podman_api.go`.
3. The CLI implementation handles both Podman and Docker binaries.

## Building

Use the provided Makefile targets:

```bash
# Format source and build the binary
make build

# Build for all supported platforms
make build-all-platforms
```

`make build` will run `go fmt`, `go mod tidy`, and `golangci-lint` before compiling.
The resulting executable is `podman-mcp-server`.

### Linting

The project uses [golangci-lint](https://golangci-lint.run/) for code quality checks:

```bash
# Run linter (automatically downloads golangci-lint if needed)
make lint
```

The linter is also run automatically as part of `make build`.

### Build Tags

The project uses several build tags for compatibility:

- **`remote`**: Use remote client mode (no local daemon required)
- **`containers_image_openpgp`**: Use pure Go OpenPGP instead of gpgme (C library)
- **`exclude_graphdriver_btrfs`**, **`btrfs_noversion`**: Exclude btrfs driver (requires C library)
- **`exclude_graphdriver_devicemapper`**: Exclude devicemapper driver (requires C library)

These tags are automatically applied by the Makefile to build, test, and lint commands.

To exclude the Podman API implementation entirely (e.g., for minimal builds):

```bash
go build -tags "exclude_podman_api" ./...
```

## Running

The README demonstrates running the server via
[`mcp-inspector`](https://modelcontextprotocol.io/docs/tools/inspector):

```bash
make build
npx @modelcontextprotocol/inspector@latest $(pwd)/podman-mcp-server
```

To run the server locally, you can use `npx`, `uvx` or execute the binary directly:

```bash
# Using npx (Node.js package runner)
npx -y podman-mcp-server@latest

# Using uvx (Python package runner)
uvx podman-mcp-server@latest

# Binary execution
./podman-mcp-server
```

### Transport Modes

The server supports multiple transport modes:

1. **STDIO mode** (default) - Communicates via standard input/output
2. **HTTP mode** (`--port`) - Modern HTTP transport with both Streamable HTTP and SSE endpoints
3. **SSE-only mode** (`--sse-port`) - Legacy Server-Sent Events transport (deprecated)

```bash
# Run with HTTP transport on a specific port (Streamable HTTP at /mcp and SSE at /sse)
./podman-mcp-server --port 8080
```

The HTTP mode uses the official MCP Go SDK's `StreamableHTTPHandler` for stateless HTTP-based communication at the `/mcp` endpoint and `SSEHandler` at the `/sse` endpoint.

**Deprecated flags:** The `--sse-port` and `--sse-base-url` flags are deprecated. Use `--port` instead, which provides both Streamable HTTP and SSE endpoints.

## Tests

Run all Go tests with:

```bash
make test
```

### Testing Philosophy

This project follows these testing principles:

1. **Black-box Testing**: Tests verify behavior and observable outcomes, not implementation details. Test the public API (MCP tools) through the test client.

2. **Real CLI with Mock Backend**: Instead of mocking the podman CLI, tests use the real podman binary pointing to a mock HTTP server. This provides realistic testing of the full CLI-to-API pipeline.

3. **Nested Test Structure**: Use `s.Run()` subtests for related scenarios within a single test function. This provides clear organization and focused failure identification.

4. **Scenario-Based Setup**: Set up mock responses before calling the tool, then verify both the result and that the expected API calls were made.

5. **Single Assertion Per Subtest**: Each `s.Run()` block should assert ONE specific condition for clear failure identification.

### Testing Patterns and Guidelines

Tests use `testify/suite` following the kubernetes-mcp-server patterns.

#### McpSuite

All tests use `McpSuite`, which runs the real podman CLI against a mock HTTP server that simulates the Podman REST API. This provides realistic testing of the full CLI-to-API pipeline.

```go
package mcp_test

import (
    "testing"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/stretchr/testify/suite"

    "github.com/manusa/podman-mcp-server/internal/test"
)

type ContainerToolsSuite struct {
    test.McpSuite
}

func TestContainerTools(t *testing.T) {
    suite.Run(t, new(ContainerToolsSuite))
}

func (s *ContainerToolsSuite) TestContainerList() {
    // Set up mock response
    s.WithContainerList([]test.ContainerListResponse{
        {
            ID:      "abc123def456",
            Names:   []string{"test-container"},
            Image:   "docker.io/library/nginx:latest",
            State:   "running",
            Status:  "Up 2 hours",
            Created: "2024-01-01T00:00:00Z",
        },
    })

    toolResult, err := s.CallTool("container_list", map[string]interface{}{})

    s.Run("returns OK", func() {
        s.NoError(err)
        s.False(toolResult.IsError)
    })

    s.Run("returns container data", func() {
        text := toolResult.Content[0].(*mcp.TextContent).Text
        s.Contains(text, "test-container")
    })

    s.Run("mock server received request", func() {
        s.True(s.MockServer.HasRequest("GET", "/libpod/containers/json"))
    })
}
```

**Note:** Tests using `McpSuite` require podman to be installed. They will fail if podman is not available.

#### Multi-Implementation Testing

Tests can be run with different configurations using the `Config` field:

```go
// Run tests with a specific implementation
func TestContainerToolsWithCLI(t *testing.T) {
    suite.Run(t, &ContainerSuite{
        McpSuite: test.McpSuite{Config: config.Config{PodmanImpl: "cli"}},
    })
}

// Run tests with all available implementations
func TestContainerSuiteWithAllImplementations(t *testing.T) {
    for _, impl := range test.AvailableImplementations() {
        suite.Run(t, &ContainerSuite{
            McpSuite: test.McpSuite{Config: config.Config{PodmanImpl: impl}},
        })
    }
}
```

Currently available implementations:
- `"cli"` - Uses podman/docker CLI (default for tests, priority 50)
- `"api"` - Uses Podman REST API via Unix socket (priority 100)

Key patterns:
- Embed `test.McpSuite` for MCP server/client setup
- Use `package mcp_test` (external test package) to avoid import cycles
- Use nested subtests with `s.Run()` for related scenarios
- Use `s.NoError()`, `s.False()`, `s.Regexp()`, `s.Contains()` for assertions
- Use `s.Require().NoError()` for setup assertions that should stop the test

#### Test Infrastructure

The `internal/test/` package provides shared test utilities:

- **`mcp.go`** - Test suite base:
  - `McpSuite` - Uses real podman CLI with mock HTTP backend
    - `Config config.Config` - Configuration for the MCP server (uses `config.Default()` if empty)
    - `CallTool()` - Call MCP tools with typed arguments
    - `CallToolRaw()` - Call MCP tools with raw JSON arguments (for testing malformed input)
    - `ListTools()` - Get list of available MCP tools
    - `WithContainerList()`, `WithContainerInspect()`, `WithContainerLogs()`
    - `WithContainerCreate()`, `WithContainerStart()`, `WithContainerStop()`
    - `WithContainerRemove()`, `WithContainerRun()`, `WithContainerWait()`
    - `WithImageList()`, `WithImagePull()`, `WithImagePush()`, `WithImageRemove()`, `WithImageBuild()`
    - `WithNetworkList()`, `WithVolumeList()`
    - `WithError()` - Inject error responses
    - `MockServer.HasRequest()` - Verify API calls were made
    - `GetCapturedRequest()` - Retrieve first captured request details for assertions
    - `PopLastCapturedRequest()` - Retrieve and remove last captured request (for multiple subtests)
  - `AvailableImplementations()` - Returns list of implementations available for testing
  - `DefaultImplementation()` - Returns the default implementation name ("cli")

- **`env.go`** - Environment utilities:
  - `RestoreEnv()` - Restore original environment variables after test

- **`mock_server.go`** - Mock Podman API server:
  - `MockPodmanServer` - HTTP test server simulating Podman REST API
  - Supports Libpod (`/libpod/...`) endpoints
  - Handles API version prefixes (`/v5.x/...`)
  - Captures request bodies for test assertions

- **`types.go`** - API response types compatible with the Libpod API

- **`podman.go`** - Podman binary helpers:
  - `IsPodmanAvailable()` - Check if real podman is installed
  - `WithContainerHost(t, url)` - Set CONTAINER_HOST for mock server

- **`helpers.go`** - General utilities:
  - `Must[T](v, err)` - Panic on error helper
  - `ReadFile(path...)` - Read file relative to caller
  - `CreateTempContainerfile(t)` - Create temporary Containerfile for build tests

#### Test Structure

Tests are organized in `pkg/mcp/` with external test packages:
- `mcp_test.go` - MCP server tests and snapshot testing
- `podman_container_test.go` - Container tool tests (`ContainerToolsSuite`)
- `podman_image_test.go` - Image tool tests (`ImageToolsSuite`)
- `podman_network_test.go` - Network tool tests (`NetworkToolsSuite`)
- `podman_volume_test.go` - Volume tool tests (`VolumeToolsSuite`)

#### Snapshot Testing

Tool definitions are snapshot tested to detect unintended changes:

```go
func (s *McpServerSuite) TestToolDefinitionsSnapshot() {
    tools, err := s.ListTools()
    s.Require().NoError(err)
    s.assertJsonSnapshot("tool_definitions.json", tools.Tools)
}
```

Snapshots are stored in `pkg/mcp/testdata/`. To update snapshots:

```bash
make test-update-snapshots
# or
UPDATE_SNAPSHOTS=1 go test ./...
```

#### Writing Tests

When adding tests:
1. Create a suite struct embedding `test.McpSuite`.
2. Set up mock responses with `s.WithContainerList()`, `s.WithImageList()`, etc.
3. Use `s.CallTool()` to invoke MCP tools.
4. Verify API calls with `s.MockServer.HasRequest()`.
5. Use `s.GetCapturedRequest()` to verify request details (query params, body).
6. Use `s.WithError()` to test error scenarios.
7. Test both success and error scenarios.

## Dependencies

When introducing new modules run `make tidy` so that `go.mod` and `go.sum` remain tidy.

Key dependencies:
- **`github.com/modelcontextprotocol/go-sdk`** - Official MCP Go SDK for production server and test client
- **`github.com/containers/podman/v5`** - Official Podman Go bindings for REST API implementation
- **`github.com/spf13/cobra`** - CLI framework
- **`github.com/spf13/viper`** - Configuration management
- **`github.com/stretchr/testify`** - Testing framework with suite support

## Coding style

- Go modules target Go **1.25** (see `go.mod`).
- Tests use `testify/suite` with the `test.McpSuite` base (see Testing section above).
- Build and test steps are defined in the Makefile—keep them working.
- Use interfaces for abstraction (see `pkg/podman/interface.go`).
- Tool definitions use SDK-agnostic types from `pkg/api/` (see SDK Architecture section above).

## Commit Message Style

This project uses [Conventional Commits](https://www.conventionalcommits.org/). All commits must be signed off with `--signoff`.

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]

Signed-off-by: Name <email>
```

### Commit Types

| Type | Description | Semantic Version Impact |
|------|-------------|------------------------|
| `feat` | A new feature | MINOR |
| `fix` | A bug fix or performance improvement | PATCH |
| `refactor` | A code change that neither fixes a bug nor adds a feature | PATCH |
| `test` | Adding missing tests or correcting existing tests | PATCH |
| `chore` | Maintenance tasks (with scope for specifics) | PATCH |
| `revert` | Reverts a previous commit | PATCH |

### Chore Scopes

Use `chore(<scope>)` for maintenance tasks:

| Scope | Description |
|-------|-------------|
| `docs` | Documentation only changes |
| `style` | Code style changes (white-space, formatting, etc.) |
| `build` | Changes that affect the build system or external dependencies |
| `ci` | Changes to CI configuration files and scripts |
| `deps` | Dependency updates |

### Guidelines

1. **Description (subject line)**:
   - Use imperative mood ("add" not "added" or "adds")
   - Don't capitalize the first letter
   - No period at the end
   - Keep under 50 characters (max 72)

2. **Scope** (optional but recommended):
   - Use lowercase
   - Examples: `api`, `mcp`, `cli`, `podman`, `docs`

3. **Body** (optional):
   - Wrap at 72 characters
   - Explain **what** and **why**, not how
   - Use blank line to separate from subject

4. **Footer** (optional):
   - Reference issues: `Fixes #123`, `Closes #456`
   - Breaking changes: `BREAKING CHANGE: description`

### Breaking Changes

Add `!` after the type/scope and include a BREAKING CHANGE footer:

```
feat(api)!: remove deprecated endpoints

BREAKING CHANGE: The /v1/users endpoint has been removed.

Signed-off-by: Name <email>
```

### Examples

```
feat(mcp): add container exec tool
```

```
fix(podman): handle empty container list response
```

```
chore(docs): update installation instructions
```

```
chore(ci): add caching to GitHub Actions workflow
```

```
refactor(api): extract tool registration into separate module

The tool registration logic was duplicated across multiple files.
This consolidates it into a single utility for better maintainability.

Signed-off-by: Name <email>
```

## Distribution Methods

The server is distributed as a binary executable, an npm package, a Python package, and is registered in the official MCP Registry.

- **Native binaries** for Linux, macOS, and Windows are available in the GitHub releases.
- An **npm** package is available at [npmjs.com](https://www.npmjs.com/package/podman-mcp-server).
  It wraps the platform-specific binary and provides a convenient way to run the server using `npx`.
- A **Python** package is available at [pypi.org](https://pypi.org/project/podman-mcp-server/).
  It provides a script that downloads the correct platform binary from the GitHub releases page and runs it.
  It provides a convenient way to run the server using `uvx` or `python -m podman_mcp_server`.
- The **MCP Registry** entry is available at [modelcontextprotocol.io](https://modelcontextprotocol.io).
  The `server.json` manifest file in the repository root defines the server metadata for the registry.

### NPM Package Structure

The npm distribution uses a modular package structure:
- `podman-mcp-server` - Main package with the wrapper script (`bin/index.js`)
- `podman-mcp-server-{os}-{arch}` - Platform-specific packages containing the binary

The `package.json` files are generated dynamically during the build process by `build/node.mk`.
Only the `bin/index.js` wrapper script is stored in the repository.

The wrapper script (`npm/podman-mcp-server/bin/index.js`):
- Resolves the correct platform-specific binary using `optionalDependencies`
- Uses `spawn` (not `execFileSync`) for proper signal handling
- Forwards SIGTERM, SIGINT, SIGHUP signals to the child process
- Returns correct exit codes based on termination signals

### Publishing Packages

Use the modular build targets for publishing:

```bash
# Publish to npm (builds all platforms first)
make npm-publish

# Publish to PyPI
make python-publish
```

The `GIT_TAG_VERSION` variable (derived from git tags) is used for package versions.

### MCP Registry Publishing

The server is automatically published to the official MCP Registry after each release.
The `.github/workflows/release-mcp-registry.yaml` workflow:
- Triggers after the Release workflow completes successfully
- Can be manually triggered with a tag input (e.g., `v0.0.51`)
- Updates the version in `server.json` (stripping the `v` prefix from the tag)
- Uses GitHub OIDC authentication with `mcp-publisher` CLI

The `server.json` manifest includes:
- Server name: `io.github.manusa/podman-mcp-server`
- npm and PyPI package definitions with STDIO transport
- Version placeholder (`0.0.0`) replaced at publish time

## Available MCP Tools

The server provides 13 tools organized by resource type:

### Container Tools
- `container_inspect` - Inspect a container's configuration and state
- `container_list` - List containers (optionally include stopped)
- `container_logs` - Get container logs
- `container_remove` - Remove a container
- `container_run` - Run a new container from an image
- `container_stop` - Stop a running container

### Image Tools
- `image_build` - Build an image from a Containerfile/Dockerfile
- `image_list` - List images
- `image_pull` - Pull an image from a registry
- `image_push` - Push an image to a registry
- `image_remove` - Remove an image

### Network Tools
- `network_list` - List networks

### Volume Tools
- `volume_list` - List volumes

## Feature Specifications

Feature specs in `docs/specs/` are **living documentation** that describe implemented features. Unlike ADRs (which are point-in-time decisions), specs are updated whenever the feature changes.

### Purpose

Specs serve as the authoritative reference for:
- **Requirements**: What the feature must do (testable statements)
- **API Contracts**: Endpoints, request/response formats, error codes
- **Architecture**: Data structures, component relationships, timing
- **Configuration**: Environment variables, constants, thresholds

### When to Read Specs

**Before modifying a feature**: Read its spec to understand current behavior, requirements, and constraints. The spec tells you what invariants must be preserved.

**Before implementing related features**: Specs document integration points and dependencies.

### When to Update Specs

**After changing a feature**: If you modify behavior, API contracts, timing, or configuration, update the spec to match. The spec must always reflect the current implementation.

**After adding requirements**: New requirements discovered during implementation should be documented.

### Available Specs

| Feature | Spec | Status | Covers |
|---------|------|--------|--------|
| Podman Interface | `docs/specs/podman-interface.md` | Implemented | `Podman` interface definition, implementation registry, `--podman-impl` flag |
| Podman REST API Bindings | `docs/specs/podman-rest-api-bindings.md` | Implemented | `api` implementation using `pkg/bindings` via Unix socket |

## Relationship to kubernetes-mcp-server

This project shares patterns and architecture with [kubernetes-mcp-server](https://github.com/containers/kubernetes-mcp-server).
When making architectural decisions, refer to kubernetes-mcp-server as the reference implementation for:
- Toolset-based architecture patterns
- Configuration system design
- Testing infrastructure (testify/suite, snapshot testing)
- Build system organization
- Documentation structure
