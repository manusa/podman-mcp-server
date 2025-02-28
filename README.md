# Podman MCP Server

[![GitHub License](https://img.shields.io/github/license/manusa/podman-mcp-server)](https://github.com/manusa/podman-mcp-server/blob/main/LICENSE)
[![npm](https://img.shields.io/npm/v/podman-mcp-server)](https://www.npmjs.com/package/podman-mcp-server)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/manusa/podman-mcp-server?sort=semver)](https://github.com/manusa/podman-mcp-server/releases/latest)
[![Build](https://github.com/manusa/podman-mcp-server/actions/workflows/build.yaml/badge.svg)](https://github.com/manusa/podman-mcp-server/actions/workflows/build.yaml)

[✨ Features](#features) | [🚀 Getting Started](#getting-started) | [🎥 Demos](#demos)

## ✨ Features <a id="features"></a>

A powerful and flexible MCP server for container runtimes supporting Podman and Docker.


## 🚀 Getting Started <a id="getting-started"></a>

### Claude Desktop

#### Using npx

If you have npm installed, this is the fastest way to get started with `podman-mcp-server` on Claude Desktop.

Open your `claude_desktop_config.json` and add the mcp server to the list of `mcpServers`:
``` json
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

### Goose CLI

[Goose CLI](https://blog.marcnuri.com/goose-on-machine-ai-agent-cli-introduction) is the easiest (and cheapest) way to get rolling with artificial intelligence (AI) agents.

#### Using npm

If you have npm installed, this is the fastest way to get started with `podman-mcp-server`.

Open your goose `config.yaml` and add the mcp server to the list of `mcpServers`:
```yaml
extensions:
  podman:
    command: npx
    args:
      - -y
      - podman-mcp-server@latest

```

## 🎥 Demos <a id="demos"></a>

