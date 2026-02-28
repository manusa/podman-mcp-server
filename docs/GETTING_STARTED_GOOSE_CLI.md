# Using with Goose CLI

[Goose CLI](https://block.github.io/goose/) is an open source AI agent that supports MCP servers as "extensions." It runs locally and connects to various LLM providers.

## Config File Location

Goose stores configuration in:

| Platform          | Path                                                                                                    |
| ----------------- | ------------------------------------------------------------------------------------------------------- |
| **macOS / Linux** | `~/.config/goose/config.yaml`                                                                           |
| **Windows**       | `%APPDATA%\Block\goose\config\config.yaml` (e.g. `C:\Users\<user>\AppData\Roaming\Block\goose\config\`) |

## Installation

### Method 1: Direct Config Edit

Add the Podman MCP server to the `extensions` section of your `config.yaml`:

```yaml
extensions:
  podman:
    name: podman
    cmd: npx
    args: [-y, podman-mcp-server@latest]
    enabled: true
    type: stdio
    envs: {}
    timeout: 300
```

If you already have other extensions, add the `podman` entry alongside them.

### Method 2: Interactive Wizard

Run the configuration wizard:

```bash
goose configure
```

1. Select **Add Extension**
2. Choose **Command-line Extension** (for local STDIO servers)
3. Follow the prompts:
   - **Name:** `podman`
   - **Command:** `npx -y podman-mcp-server@latest`
   - **Timeout:** `300` (recommended for slower startup; increase if needed)
   - **Environment variables:** skip unless required
4. Save and exit

Example wizard output:

```
┌   goose-configure
│
◇  What would you like to configure?
│  Add Extension
│
◇  What type of extension would you like to add?
│  Command-line Extension
│
◇  What would you like to call this extension?
│  podman
│
◇  What command should be run?
│  npx -y podman-mcp-server@latest
│
◇  Please set the timeout for this tool (in secs):
│  300
│
◆  Would you like to add environment variables?
│  No
│
└  Added podman extension
```

## Requirements

- [Node.js](https://nodejs.org/) (for `npx`) or the [standalone binary](https://github.com/manusa/podman-mcp-server/releases/latest)
- Podman or Docker installed and running (see [Getting Started with Podman/Docker](GETTING_STARTED_PODMAN.md))

## Verifying

1. **Check extension is enabled:**

   ```bash
   goose configure
   ```

   Select **Toggle Extensions** and confirm `podman` appears and is enabled.

2. **Start a session and discover tools:**

   ```bash
   goose session
   ```

   Ask: "What tools do you have?" — Podman tools should be listed.

3. **Run a simple task:** Ask Goose to list containers, pull an image, or run a container.

## Troubleshooting

- **Command not found / Extension fails:** Ensure `npx` is on your PATH. Use the full path to the binary if needed, or run `npm install -g npm` to ensure npm/npx are available.
- **Timeouts:** Increase the extension `timeout` (e.g. 300–600 seconds) in `config.yaml` or via `goose configure`.
- **Missing logs:** Create `~/.config/goose/logs` manually and ensure the config directory is writable.
- **Tools not discovered:** Run `goose session` and ask "What tools do you have?" to verify. Check that Podman or Docker is running.

## References

- [Goose Documentation](https://block.github.io/goose/docs/) – Official docs
- [Using Extensions](https://block.github.io/goose/docs/getting-started/using-extensions/) – Extension configuration guide