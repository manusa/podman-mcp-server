package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initPodmanVolume() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("volume_list",
			mcp.WithDescription("List all the available Docker or Podman volumes"),
		), s.volumeList},
	}
}

func (s *Server) volumeList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.VolumeList()), nil
}
