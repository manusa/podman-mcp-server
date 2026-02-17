# Using with VS Code / Cursor

Configure the Podman MCP Server with VS Code or Cursor.

## Quick Install

**VS Code Insiders** – Click to install:

[<img src="https://img.shields.io/badge/VS_Code-VS_Code?style=flat-square&label=Install%20Server&color=0098FF" alt="Install in VS Code">](https://insiders.vscode.dev/redirect?url=vscode%3Amcp%2Finstall%3F%257B%2522name%2522%253A%2522podman%2522%252C%2522command%2522%253A%2522npx%2522%252C%2522args%2522%253A%255B%2522-y%2522%252C%2522podman-mcp-server%2540latest%2522%255D%257D)
[<img alt="Install in VS Code Insiders" src="https://img.shields.io/badge/VS_Code_Insiders-VS_Code_Insiders?style=flat-square&label=Install%20Server&color=24bfa5">](https://insiders.vscode.dev/redirect?url=vscode-insiders%3Amcp%2Finstall%3F%257B%2522name%2522%253A%2522podman%2522%252C%2522command%2522%253A%2522npx%2522%252C%2522args%2522%253A%255B%2522-y%2522%252C%2522podman-mcp-server%2540latest%2522%255D%257D)

## Manual Install

Run one of the following commands:

```shell
# For VS Code
code --add-mcp '{"name":"podman","command":"npx","args":["podman-mcp-server@latest"]}'

# For VS Code Insiders
code-insiders --add-mcp '{"name":"podman","command":"npx","args":["podman-mcp-server@latest"]}'

# For Cursor
cursor --add-mcp '{"name":"podman","command":"npx","args":["podman-mcp-server@latest"]}'
```

## Requirements

- [Node.js](https://nodejs.org/) (for `npx`) or the [standalone binary](https://github.com/manusa/podman-mcp-server/releases/latest)
- Podman or Docker installed and running (see [Getting Started with Podman/Docker](GETTING_STARTED_PODMAN.md))

## Verifying

After installation, the Podman MCP server tools should be available in your AI chat. Try asking to list containers or images.
