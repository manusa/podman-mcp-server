# Getting Started with Podman/Docker

This guide covers the prerequisites for running the Podman MCP Server.

## Prerequisites

The Podman MCP Server requires either **Podman** or **Docker** to be installed and available on your system.

### Podman

Install Podman for your platform:

- **Linux & macOS**: [Podman Installation](https://podman.io/docs/installation)
- **Windows**: [Podman Desktop](https://podman-desktop.io/) or WSL2 with Podman

Verify installation:

```shell
podman --version
podman info
```

### Docker

If you prefer Docker, ensure it is installed and the daemon is running:

```shell
docker --version
docker info
```

The server automatically detects and uses whichever runtime (Podman or Docker) is available. Podman is preferred when both are present.

## Next Steps

Once your container runtime is ready, proceed to configure the MCP server with your AI assistant:

- [Using with Claude Desktop](GETTING_STARTED_CLAUDE_DESKTOP.md)
- [Using with VS Code / Cursor](GETTING_STARTED_VSCODE.md)
- [Using with Goose CLI](GETTING_STARTED_GOOSE.md)
