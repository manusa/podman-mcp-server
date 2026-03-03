# Using with Claude Desktop

Configure the Podman MCP Server with [Claude Desktop](https://claude.ai/download). Claude Desktop is available for macOS and Windows. Linux users should use [Claude Code CLI](getting-started-claude-code.md) instead.

## Config File Location

The `claude_desktop_config.json` file location varies by operating system:

| Platform    | Path                                                                                           |
| ----------- | ---------------------------------------------------------------------------------------------- |
| **macOS**   | `~/Library/Application Support/Claude/claude_desktop_config.json`                              |
| **Windows** | `%APPDATA%\Claude\claude_desktop_config.json` (e.g. `C:\Users\<user>\AppData\Roaming\Claude\`) |

### Opening the Config File

Claude menu → Settings... → Developer → Edit Config. This creates the file if it doesn't exist.

## Installation

Add the Podman MCP server to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "podman": {
      "command": "npx",
      "args": [
        "-y",
        "podman-mcp-server@latest"
      ]
    }
  }
}
```

### Alternative: Using uvx

If you have Python installed instead of Node.js:

```json
{
  "mcpServers": {
    "podman": {
      "command": "uvx",
      "args": [
        "podman-mcp-server@latest"
      ]
    }
  }
}
```

## Requirements

- [Node.js](https://nodejs.org/) (for `npx`) or [Python](https://www.python.org/) (for `uvx`) or the [standalone binary](https://github.com/manusa/podman-mcp-server/releases/latest)
- Podman or Docker installed and running (see [Getting Started with Podman/Docker](getting-started-podman.md))

## Alternative: Standalone Binary

If you prefer not to use npm, download the binary for your platform and configure the path:

**macOS:**

```json
{
  "mcpServers": {
    "podman": {
      "command": "/path/to/podman-mcp-server"
    }
  }
}
```

**Windows:** Use forward slashes or escaped backslashes:

```json
{
  "mcpServers": {
    "podman": {
      "command": "C:/path/to/podman-mcp-server.exe"
    }
  }
}
```

## Verifying

1. **Restart Claude Desktop completely** (closing the window is not enough).
2. Look for the MCP server indicator in the bottom-right of the chat input — this indicates MCP servers are loaded.
3. Click the indicator to see available tools, including Podman tools.
4. Ask Claude to list containers, pull images, or run containers.

## Troubleshooting

- **Server not showing:** Check [Claude Desktop logs](https://modelcontextprotocol.io/docs/develop/connect-local-servers#getting-logs-from-claude-for-desktop): `mcp.log` and `mcp-server-podman.log` in `~/Library/Logs/Claude` (macOS) or `%APPDATA%\Claude\logs` (Windows).
- **Invalid JSON:** A single trailing comma or syntax error will silently break the config. Validate your JSON.
- **Manual test:** Run `npx -y podman-mcp-server@latest` in a terminal to verify the server starts.

## References

- [Connect to local MCP servers](https://modelcontextprotocol.io/docs/develop/connect-local-servers) — Official MCP documentation
- [Claude Desktop download](https://claude.ai/download)
