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
- `.github/` – GitHub-related configuration (Actions workflows, Dependabot).
- `npm/` – Node packages that wrap the compiled binaries for distribution through npmjs.com.
- `python/` – Python package providing a script that downloads the correct platform binary from the GitHub releases page and runs it for distribution through pypi.org.
- `testdata/` – test fixtures including a mock podman binary for testing.
- `Makefile` – tasks for building, formatting, and testing.

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

> **Note:** Tests are currently written with vanilla `testing` package but should be migrated to `testify/suite`
> to match kubernetes-mcp-server patterns. See [Issue #68](https://github.com/manusa/podman-mcp-server/issues/68).

#### Target Testing Pattern (testify/suite)

New tests should use `testify/suite` following the kubernetes-mcp-server pattern:

```go
type ContainerToolsSuite struct {
    suite.Suite
    // test fixtures
}

func (s *ContainerToolsSuite) SetupTest() {
    // setup before each test
}

func (s *ContainerToolsSuite) TestContainerList() {
    s.Run("returns empty list when no containers", func() {
        // test implementation
        s.Equal(expected, actual)
    })
    s.Run("returns all containers when requested", func() {
        // test implementation
    })
}

func TestContainerTools(t *testing.T) {
    suite.Run(t, new(ContainerToolsSuite))
}
```

Key patterns:
- Use `testify/suite` for organizing tests into suites
- Use nested subtests with `s.Run()` for related scenarios
- Use `s.Equal()`, `s.True()`, `s.NoError()` for assertions
- One assertion per test case when possible

#### Current Testing Infrastructure

This project currently uses a mock Podman binary approach for testing:

- The `testdata/podman/main.go` file contains a fake podman binary that:
  - Echoes the command line arguments it receives
  - Optionally reads from an `output.txt` file to provide mock responses
- Tests build this fake binary and prepend its directory to PATH
- This allows testing MCP tools without requiring an actual Podman installation

#### Test Structure

Tests are organized in `pkg/mcp/` alongside the source files:
- `common_test.go` - Shared test infrastructure (mock binary setup, MCP client helpers)
- `mcp_test.go` - MCP server integration tests
- `podman_container_test.go` - Container tool tests
- `podman_image_test.go` - Image tool tests
- `podman_network_test.go` - Network tool tests
- `podman_volume_test.go` - Volume tool tests

#### Test Helpers

The `mcpContext` struct in `common_test.go` provides:
- `testCase()` - Sets up test environment with mock binary and MCP client
- `callTool()` - Calls an MCP tool with arguments
- `withPodmanOutput()` - Injects mock output for podman commands

#### Writing Tests

When adding tests:
1. Use `testify/suite` pattern for new test files.
2. Use the `testCase()` helper to set up the test environment.
3. Use `withPodmanOutput()` to inject expected command output.
4. Call tools using `callTool()` and verify the results.
5. Test both success and error scenarios.
6. Use nested subtests with `s.Run()` for related scenarios.

## Dependencies

When introducing new modules run `make tidy` so that `go.mod` and `go.sum` remain tidy.

## Coding style

- Go modules target Go **1.24** (see `go.mod`).
- Tests are written with the standard library `testing` package.
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
