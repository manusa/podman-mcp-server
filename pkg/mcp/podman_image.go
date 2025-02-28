package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initPodmanImage() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("container_image_pull",
			mcp.WithDescription("Copies (pulls) a Docker or Podman container image from a registry onto the local machine"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to pull"), mcp.Required()),
		), s.imagePull},
	}
}

func (s *Server) imagePull(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImagePull(ctr.Params.Arguments["imageName"].(string))), nil
}
