# Podman MCP Server

[![GitHub License](https://img.shields.io/github/license/manusa/podman-mcp-server)](https://github.com/manusa/podman-mcp-server/blob/main/LICENSE)
[![npm](https://img.shields.io/npm/v/podman-mcp-server)](https://www.npmjs.com/package/podman-mcp-server)
[![PyPI - Version](https://img.shields.io/pypi/v/podman-mcp-server)](https://pypi.org/project/podman-mcp-server/)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/manusa/podman-mcp-server?sort=semver)](https://github.com/manusa/podman-mcp-server/releases/latest)
[![Build](https://github.com/manusa/podman-mcp-server/actions/workflows/build.yaml/badge.svg)](https://github.com/manusa/podman-mcp-server/actions/workflows/build.yaml)

[✨ Features](#features) | [🚀 Getting Started](#getting-started) | [📚 Documentation](#documentation) | [🎥 Demos](#demos) | [⚙️ Configuration](#configuration) | [🛠️ Tools](#tools) | [🧑‍💻 Development](#development)

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

## 📚 Documentation <a id="documentation"></a>

For detailed setup guides, configuration reference, and feature specifications, see the **[Documentation](docs/README.md)**.

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

| Option           | Description                                                                                      |
|------------------|--------------------------------------------------------------------------------------------------|
| `--port`, `-p`   | Starts the MCP server in HTTP mode with Streamable HTTP at `/mcp` and SSE at `/sse` endpoints.  |
| `--sse-port`     | **Deprecated.** Use `--port` instead. Starts the MCP server in SSE-only mode.                   |
| `--sse-base-url` | **Deprecated.** SSE public base URL to use when sending the endpoint message.                   |

### Transport Modes

The server supports multiple transport modes:

1. **STDIO mode** (default) - Communicates via standard input/output
2. **HTTP mode** (`--port`) - Modern HTTP transport with both Streamable HTTP and SSE endpoints
3. **SSE-only mode** (`--sse-port`) - Legacy Server-Sent Events transport (deprecated)

```shell
# Start HTTP server on port 8080 (Streamable HTTP at /mcp and SSE at /sse)
podman-mcp-server --port 8080

# Legacy SSE-only server on port 8080 (deprecated, use --port instead)
podman-mcp-server --sse-port 8080
```

## 🛠️ Tools <a id="tools"></a>

<!-- AVAILABLE-TOOLS-START -->

<details>

<summary>Container</summary>

- **container_inspect** - Displays the low-level information and configuration of a Docker or Podman container with the specified container ID or name
  - `name` (`string`) **(required)** - Docker or Podman container ID or name to display the information

- **container_list** - Prints out information about the running Docker or Podman containers

- **container_logs** - Displays the logs of a Docker or Podman container with the specified container ID or name
  - `name` (`string`) **(required)** - Docker or Podman container ID or name to display the logs

- **container_remove** - Removes a Docker or Podman container with the specified container ID or name (rm)
  - `name` (`string`) **(required)** - Docker or Podman container ID or name to remove

- **container_run** - Runs a Docker or Podman container with the specified image name
  - `environment` (`array`) - Environment variables to set in the container. Format: <key>=<value>. Example: FOO=bar. (Optional, add only to set environment variables)
  - `imageName` (`string`) **(required)** - Docker or Podman container image name to run
  - `ports` (`array`) - Port mappings to expose on the host. Format: <hostPort>:<containerPort>. Example: 8080:80. (Optional, add only to expose ports)

- **container_stop** - Stops a Docker or Podman running container with the specified container ID or name
  - `name` (`string`) **(required)** - Docker or Podman container ID or name to stop

</details>

<details>

<summary>Image</summary>

- **image_build** - Build a Docker or Podman image from a Dockerfile, Podmanfile, or Containerfile
  - `containerFile` (`string`) **(required)** - The absolute path to the Dockerfile, Podmanfile, or Containerfile to build the image from
  - `imageName` (`string`) - Specifies the name which is assigned to the resulting image if the build process completes successfully (--tag, -t)

- **image_list** - List the Docker or Podman images on the local machine

- **image_pull** - Copies (pulls) a Docker or Podman container image from a registry onto the local machine storage
  - `imageName` (`string`) **(required)** - Docker or Podman container image name to pull

- **image_push** - Pushes a Docker or Podman container image, manifest list or image index from local machine storage to a registry
  - `imageName` (`string`) **(required)** - Docker or Podman container image name to push

- **image_remove** - Removes a Docker or Podman image from the local machine storage
  - `imageName` (`string`) **(required)** - Docker or Podman container image name to remove

</details>

<details>

<summary>Network</summary>

- **network_list** - List all the available Docker or Podman networks

</details>

<details>

<summary>Volume</summary>

- **volume_list** - List all the available Docker or Podman volumes

</details>


<!-- AVAILABLE-TOOLS-END -->

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
