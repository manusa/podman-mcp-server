# Project Agents.md for Podman MCP Server

This Agents.md file provides comprehensive guidance for AI assistants and coding agents (like Claude, Gemini, Cursor, and others) to work with this codebase.

This repository contains the podman-mcp-server project,
a Go-based Model Context Protocol (MCP) server that provides container management capabilities using Podman or Docker.
This MCP server enables AI assistants (like Claude, Gemini, Cursor, and others) to interact with container runtimes using the Model Context Protocol (MCP).

## Project Structure and Repository layout

- Go package layout follows the standard Go conventions:
  - `cmd/podman-mcp-server/` – main application entry point.
  - `pkg/` – libraries grouped by domain.
    - `mcp/` - Model Context Protocol (MCP) server implementation with tool definitions for containers, images, networks, and volumes.
    - `podman/` - Podman/Docker CLI abstraction layer with interface definition and CLI implementation.
    - `podman-mcp-server/cmd/` - CLI command definition using Cobra framework.
    - `version/` - Version information management.
  - `internal/test/` – shared test utilities (McpSuite, mock podman helpers).
- `.github/` – GitHub-related configuration (Actions workflows, Dependabot).
- `build/` – modular Makefile includes for packaging targets.
  - `node.mk` – NPM packaging targets (npm-copy-binaries, npm-copy-project-files, npm-publish).
  - `python.mk` – Python/PyPI packaging targets (python-publish).
- `npm/` – Node packages that wrap the compiled binaries for distribution through npmjs.com.
- `python/` – Python package providing a script that downloads the correct platform binary from the GitHub releases page and runs it for distribution through pypi.org.
- `testdata/` – test fixtures including a mock podman binary for testing.
- `Makefile` – core build tasks; includes `build/*.mk` for packaging targets.

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
2. Define the tool using `mcp.NewTool()` with appropriate parameters and description.
3. Implement the handler function that executes the tool's logic.
4. Add the tool and handler to the `tools()` and `handlers()` functions in the resource file.
5. Register the tool in `mcp.go` by importing it in the `initTools()` function.
6. Add tests for the new tool.

### Podman Interface

The `pkg/podman/interface.go` file defines the `Podman` interface that abstracts container runtime operations.
The `pkg/podman/podman_cli.go` file implements this interface using the Podman/Docker CLI.

When adding new container operations:
1. Add the method signature to the `Podman` interface in `interface.go`.
2. Implement the method in `podman_cli.go`.
3. The CLI implementation handles both Podman and Docker binaries.

## Building

Use the provided Makefile targets:

```bash
# Format source and build the binary
make build

# Build for all supported platforms
make build-all-platforms
```

`make build` will run `go fmt` and `go mod tidy` before compiling.
The resulting executable is `podman-mcp-server`.

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

The server supports two transport modes:

1. **STDIO mode** (default) - communicates via standard input/output
2. **SSE mode** - Server-Sent Events over HTTP

```bash
# Run with SSE transport on a specific port
./podman-mcp-server --sse-port 8080

# Run with custom base URL for SSE
./podman-mcp-server --sse-port 8080 --sse-base-url http://localhost:8080
```

## Tests

Run all Go tests with:

```bash
make test
```

### Testing Patterns and Guidelines

Tests use `testify/suite` following the kubernetes-mcp-server patterns.

#### Test Suites

All tests use `testify/suite` with the `test.McpSuite` base from `internal/test/`:

```go
package mcp_test

import (
    "testing"

    "github.com/mark3labs/mcp-go/mcp"
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
    toolResult, err := s.CallTool("container_list", map[string]interface{}{})
    s.Run("returns OK", func() {
        s.NoError(err)
        s.False(toolResult.IsError)
    })
    s.Run("lists all containers", func() {
        s.Regexp("^podman container list", toolResult.Content[0].(mcp.TextContent).Text)
    })
}
```

Key patterns:
- Embed `test.McpSuite` for MCP server/client setup
- Use `package mcp_test` (external test package) to avoid import cycles
- Use nested subtests with `s.Run()` for related scenarios
- Use `s.NoError()`, `s.False()`, `s.Regexp()`, `s.Contains()` for assertions
- Use `s.Require().NoError()` for setup assertions that should stop the test

#### Test Infrastructure

The `internal/test/` package provides shared test utilities:

- **`mcp.go`** - `McpSuite` base suite with:
  - `SetupTest()` / `TearDownTest()` - MCP server and client lifecycle
  - `CallTool(name, args)` - Call an MCP tool
  - `ListTools()` - List available tools
  - `WithPodmanOutput(lines...)` - Inject mock podman output

- **`podman.go`** - Mock podman binary helpers:
  - `WithPodmanBinary(t)` - Build fake podman and prepend to PATH

- **`helpers.go`** - General utilities:
  - `Must[T](v, err)` - Panic on error helper
  - `ReadFile(path...)` - Read file relative to caller

#### Mock Podman Binary

The `testdata/podman/main.go` contains a fake podman binary that:
- Echoes the command line arguments it receives
- Optionally reads from an `output.txt` file to provide mock responses

This allows testing MCP tools without requiring an actual Podman installation.

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
2. Use `s.CallTool()` to invoke MCP tools.
3. Use `s.WithPodmanOutput()` to inject expected command output.
4. Use `s.Run()` for nested subtests with related scenarios.
5. Use `s.Regexp("^pattern", text)` for prefix matching, `s.Regexp("pattern$", text)` for suffix.
6. Test both success and error scenarios.

## Dependencies

When introducing new modules run `make tidy` so that `go.mod` and `go.sum` remain tidy.

## Coding style

- Go modules target Go **1.24** (see `go.mod`).
- Tests use `testify/suite` with the `test.McpSuite` base (see Testing section above).
- Build and test steps are defined in the Makefile—keep them working.
- Use interfaces for abstraction (see `pkg/podman/interface.go`).

## Distribution Methods

The server is distributed as a binary executable, an npm package, and a Python package.

- **Native binaries** for Linux, macOS, and Windows are available in the GitHub releases.
- An **npm** package is available at [npmjs.com](https://www.npmjs.com/package/podman-mcp-server).
  It wraps the platform-specific binary and provides a convenient way to run the server using `npx`.
- A **Python** package is available at [pypi.org](https://pypi.org/project/podman-mcp-server/).
  It provides a script that downloads the correct platform binary from the GitHub releases page and runs it.
  It provides a convenient way to run the server using `uvx` or `python -m podman_mcp_server`.

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

## Relationship to kubernetes-mcp-server

This project shares patterns and architecture with [kubernetes-mcp-server](https://github.com/containers/kubernetes-mcp-server).
When making architectural decisions, refer to kubernetes-mcp-server as the reference implementation for:
- Toolset-based architecture patterns
- Configuration system design
- Testing infrastructure (testify/suite, snapshot testing)
- Build system organization
- Documentation structure
