# Using with Claude Code CLI

[Claude Code](https://docs.anthropic.com/en/docs/claude-code) is Anthropic's CLI tool for agentic coding. It supports MCP servers for extending its capabilities.

## Quick Setup

Add the MCP server using the `claude mcp add` command:

```bash
claude mcp add podman -- npx -y podman-mcp-server@latest
```

This makes the Podman MCP server available in all your Claude Code sessions.

## Alternative: Using uvx

If you have Python installed instead of Node.js:

```bash
claude mcp add podman -- uvx podman-mcp-server@latest
```

## Verify Connection

After adding the MCP server, verify it's connected:

```bash
claude mcp list
```

The `podman` server should appear in the list and show as connected.

## Requirements

- [Node.js](https://nodejs.org/) (for `npx`) or [Python](https://www.python.org/) (for `uvx`) or the [standalone binary](https://github.com/manusa/podman-mcp-server/releases/latest)
- Podman or Docker installed and running (see [Getting Started with Podman/Docker](getting-started-podman.md))

## Using the MCP Server

Once connected, Claude Code can manage containers using natural language:

```
> List all running containers

● I'll list the running containers using the Podman MCP server.
  ⎿  ID            NAMES             IMAGE                              STATE     STATUS
     abc123def456  my-nginx          docker.io/library/nginx:latest     running   Up 2 hours

● You have 1 running container: my-nginx using the nginx:latest image, running for 2 hours.
```

## Troubleshooting

- **Server not connecting:** Run `claude mcp list` to check the server status. If it shows an error, verify `npx` or `uvx` is on your PATH.
- **Manual test:** Run `npx -y podman-mcp-server@latest` in a terminal to verify the server starts.
- **Remove and re-add:** If issues persist, remove with `claude mcp remove podman` and add again.

## References

- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code) — Official docs
- [Claude Code MCP Servers](https://docs.anthropic.com/en/docs/claude-code/mcp-servers) — MCP server configuration guide
