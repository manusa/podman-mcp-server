# Podman MCP Server

[![GitHub License](https://img.shields.io/github/license/manusa/podman-mcp-server)](https://github.com/manusa/podman-mcp-server/blob/main/LICENSE)
[![npm](https://img.shields.io/npm/v/podman-mcp-server)](https://www.npmjs.com/package/podman-mcp-server)
[![PyPI - Version](https://img.shields.io/pypi/v/podman-mcp-server)](https://pypi.org/project/podman-mcp-server/)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/manusa/podman-mcp-server?sort=semver)](https://github.com/manusa/podman-mcp-server/releases/latest)
[![Build](https://github.com/manusa/podman-mcp-server/actions/workflows/build.yaml/badge.svg)](https://github.com/manusa/podman-mcp-server/actions/workflows/build.yaml)

[✨ Features](#features) | [🚀 Getting Started](#getting-started) | [🎥 Demos](#demos) | [⚙️ Configuration](#configuration) | [🧑‍💻 Development](#development)

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

### VS Code / VS Code Insiders

Install the Podman MCP server extension in VS Code Insiders by pressing the following link:

[<img src="https://img.shields.io/badge/VS_Code-VS_Code?style=flat-square&label=Install%20Server&color=0098FF" alt="Install in VS Code">](https://insiders.vscode.dev/redirect?url=vscode%3Amcp%2Finstall%3F%257B%2522name%2522%253A%2522podman%2522%252C%2522command%2522%253A%2522npx%2522%252C%2522args%2522%253A%255B%2522-y%2522%252C%2522podman-mcp-server%2540latest%2522%255D%257D)
[<img alt="Install in VS Code Insiders" src="https://img.shields.io/badge/VS_Code_Insiders-VS_Code_Insiders?style=flat-square&label=Install%20Server&color=24bfa5">](https://insiders.vscode.dev/redirect?url=vscode-insiders%3Amcp%2Finstall%3F%257B%2522name%2522%253A%2522podman%2522%252C%2522command%2522%253A%2522npx%2522%252C%2522args%2522%253A%255B%2522-y%2522%252C%2522podman-mcp-server%2540latest%2522%255D%257D)

Alternatively, you can install the extension manually by running the following command:

```shell
# For VS Code
code --add-mcp '{"name":"podman","command":"npx","args":["podman-mcp-server@latest"]}'
# For VS Code Insiders
code-insiders --add-mcp '{"name":"podman","command":"npx","args":["podman-mcp-server@latest"]}'
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

## ⚙️ Configuration <a id="configuration"></a>

The Podman MCP server can be configured using command line (CLI) arguments.

You can run the CLI executable either by using `npx` or by downloading the [latest release binary](https://github.com/manusa/podman-mcp-server/releases/latest).

```shell
# Run the Podman MCP server using npx (in case you have npm installed)
npx podman-mcp-server@latest --help
```

```shell
# Run the Podman MCP server using the latest release binary
./podman-mcp-server --help
```

### Configuration Options

| Option       | Description                                                                              |
|--------------|------------------------------------------------------------------------------------------|
| `--sse-port` | Starts the MCP server in Server-Sent Event (SSE) mode and listens on the specified port. |

## 🧑‍💻 Development <a id="development"></a>

### Running with mcp-inspector

Compile the project and run the Podman MCP server with [mcp-inspector](https://modelcontextprotocol.io/docs/tools/inspector) to inspect the MCP server.

```shell
# Compile the project
make build
# Run the Podman MCP server with mcp-inspector
npx @modelcontextprotocol/inspector@latest $(pwd)/podman-mcp-server
```

mcp-name: io.github.manusa/podman-mcp-server
