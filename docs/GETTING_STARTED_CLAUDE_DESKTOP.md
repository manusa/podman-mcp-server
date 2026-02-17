# Using with Claude Desktop

Configure the Podman MCP Server with Claude Desktop.

## Config File Location

The `claude_desktop_config.json` file location varies by operating system:

| Platform    | Path                                                                                           |
| ----------- | ---------------------------------------------------------------------------------------------- |
| **macOS**   | `~/Library/Application Support/Claude/claude_desktop_config.json`                              |
| **Windows** | `%APPDATA%\Claude\claude_desktop_config.json` (e.g. `C:\Users\<user>\AppData\Roaming\Claude\`) |
| **Linux**   | `~/.config/Claude/claude_desktop_config.json`                                                  |

### Opening the Config File

**Via Claude Desktop (macOS/Windows):** Claude menu → Settings → Developer → Edit Config. This creates the file if it doesn't exist.

**Manually:** Open the path above in your editor. On Linux, you may need to create the file and directory first:

```bash
mkdir -p ~/.config/Claude
nano ~/.config/Claude/claude_desktop_config.json
```

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

If you already have other MCP servers, add the `podman` entry inside the existing `mcpServers` object:

```json
{
  "mcpServers": {
    "filesystem": { ... },
    "podman": {
      "command": "npx",
      "args": ["-y", "podman-mcp-server@latest"]
    }
  }
}
```

## Requirements

- [Node.js](https://nodejs.org/) (for `npx`) or the [standalone binary](https://github.com/manusa/podman-mcp-server/releases/latest)
- Podman or Docker installed and running (see [Getting Started with Podman/Docker](GETTING_STARTED_PODMAN.md))

## Alternative: Standalone Binary

If you prefer not to use npm, download the binary for your platform and configure the path:

**macOS/Linux:**

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

## Platform-Specific Notes

### Windows

- If the server fails to load with `${APPDATA}`-related errors, add the expanded path to `env`:
  
  ```json
  "podman": {
    "command": "npx",
    "args": ["-y", "podman-mcp-server@latest"],
    "env": {
      "APPDATA": "C:\\Users\\<user>\\AppData\\Roaming"
    }
  }
  ```

- Ensure `npx` is on your PATH (npm is installed globally).

### Linux

- Ensure `~/.config/Claude` exists before creating the config file.
- If using a binary, make it executable: `chmod +x podman-mcp-server`.

## Verifying

1. **Restart Claude Desktop completely** (closing the window is not enough).
2. Look for the 🔨 hammer icon in the bottom-right of the chat input—this indicates MCP servers are loaded.
3. Click the icon to see available tools, including Podman tools.
4. Ask Claude to list containers, pull images, or run containers.

## Troubleshooting

- **Server not showing / hammer icon missing:** Check [Claude Desktop logs](https://modelcontextprotocol.io/docs/develop/connect-local-servers#getting-logs-from-claude-for-desktop): `mcp.log` and `mcp-server-podman.log` in `~/Library/Logs/Claude` (macOS) or `%APPDATA%\Claude\logs` (Windows). On Linux, look for logs in `~/.config/Claude/` or use Settings → Developer → Open Logs Folder if available.
- **Invalid JSON:** A single trailing comma or syntax error will silently break the config. Validate your JSON.
- **Manual test:** Run `npx -y podman-mcp-server@latest` in a terminal to verify the server starts.

## References

- [Connect to local MCP servers](https://modelcontextprotocol.io/docs/develop/connect-local-servers) – Official MCP documentation
- [Claude Desktop download](https://claude.ai/download)
