package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initPodmanNetwork() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("network_list",
			mcp.WithDescription("List all the available Docker or Podman networks"),
		), s.networkList},
	}
}

func (s *Server) networkList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.NetworkList()), nil
}
