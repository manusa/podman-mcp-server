package mcp

import (
	// Blank imports trigger each toolset's init() self-registration.
	_ "github.com/manusa/podman-mcp-server/pkg/toolsets/container"
	_ "github.com/manusa/podman-mcp-server/pkg/toolsets/image"
	_ "github.com/manusa/podman-mcp-server/pkg/toolsets/network"
	_ "github.com/manusa/podman-mcp-server/pkg/toolsets/volume"
)
