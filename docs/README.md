# Podman MCP Server Documentation

Welcome to the Podman MCP Server documentation! This directory contains guides to help you set up and use the Podman MCP Server with Podman or Docker and your preferred AI assistant client.

## Getting Started Guides

Choose the guide that matches your needs:

| Guide                                                               | Description                                                               | Best For                        |
| ------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------- |
| **[Getting Started with Podman/Docker](getting-started-podman.md)** | Prerequisites: Install Podman or Docker and verify the runtime is working | Everyone – **start here first** |
| **[Using with Claude Desktop](getting-started-claude-desktop.md)**  | Configure the MCP server with Claude Desktop                              | Claude Desktop users            |
| **[Using with Claude Code CLI](getting-started-claude-code.md)**    | Configure the MCP server with Claude Code CLI                             | Claude Code CLI users           |
| **[Using with VS Code / Cursor](getting-started-vscode.md)**        | Configure the MCP server with VS Code or Cursor                           | VS Code and Cursor users        |
| **[Using with Goose CLI](getting-started-goose-cli.md)**            | Configure the MCP server with Goose CLI                                   | Goose CLI users                 |

## Recommended Workflow

1. **Complete the base setup**: Start with [Getting Started with Podman/Docker](getting-started-podman.md) to ensure Podman or Docker is installed and accessible.
2. **Configure your client**: Then follow the guide for your preferred AI assistant (Claude Desktop, Claude Code, VS Code/Cursor, or Goose CLI).

## Feature Specifications

Living documentation for implemented and planned features:

| Spec                                                              | Description                                                         | Status      |
| ----------------------------------------------------------------- | ------------------------------------------------------------------- | ----------- |
| **[Podman Interface](specs/podman-interface.md)**                 | Interface definition, implementation registry, `--podman-impl` flag | Implemented |
| **[Podman REST API Bindings](specs/podman-rest-api-bindings.md)** | API implementation using `pkg/bindings` via Unix socket              | Implemented |

## Additional Documentation

- **[Main README](../README.md)** – Project overview, installation, and quick start
- **[AGENTS.md](../AGENTS.md)** – Guidance for AI assistants and coding agents working with this codebase
