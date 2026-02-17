# Using with Goose CLI

[Goose CLI](https://block.github.io/goose/) is an open source AI agent that supports MCP servers as "extensions." It runs locally and connects to various LLM providers.

## Config File Location

Goose stores configuration in:

| Platform          | Path                                                                                                    |
| ----------------- | ------------------------------------------------------------------------------------------------------- |
| **macOS / Linux** | `~/.config/goose/config.yaml`                                                                           |
| **Windows**       | `%APPDATA%\Block\goose\config\config.yaml` (e.g. `C:\Users\<user>\AppData\Roaming\Block\goose\config\`) |

Related files in the same directory: `profiles.yaml` (providers and extension settings). Create `~/.config/goose/logs` manually if logs are missing.

## Installation

### Method 1: Direct Config Edit

Add the Podman MCP server to the `extensions` section of your `config.yaml`:

```yaml
extensions:
  podman:
    enabled: true
    type: stdio
    name: podman
    description: this is a mcp server
    cmd: npx
    args:
    - -y
    - podman-mcp-server@latest
    envs: {}
    env_keys: []
    timeout: 300
    bundled: null
    available_tools: []
```

If you already have other extensions, add the `podman` entry alongside them.

### Method 2: Interactive Wizard

Run the configuration wizard:

```bash
goose configure
```

1. Select **Add Extension**
2. Choose **Command-line Extension** (for local STDIO servers)
3. Set:
   - **Name:** `podman`
   - **Command:** `npx -y podman-mcp-server@latest`
   - **Timeout:** `300` (recommended for slower startup; increase if needed)
4. Skip environment variables unless required
5. Save and exit

## Requirements

- [Node.js](https://nodejs.org/) (for `npx`) or the [standalone binary](https://github.com/manusa/podman-mcp-server/releases/latest)
- Podman or Docker installed and running (see [Getting Started with Podman/Docker](GETTING_STARTED_PODMAN.md))

## Verifying

1. **Check extension is recorded:**

   ```bash
   goose info -v
   ```

   Confirm `podman` appears under extensions.

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
