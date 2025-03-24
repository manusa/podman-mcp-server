package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initPodmanImage() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("image_list",
			mcp.WithDescription("List the Docker or Podman images on the local machine"),
		), s.containerImageList},
		{mcp.NewTool("image_pull",
			mcp.WithDescription("Copies (pulls) a Docker or Podman container image from a registry onto the local machine"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to pull"), mcp.Required()),
		), s.containerImagePull},
	}
}

func (s *Server) containerImageList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImageList()), nil
}

func (s *Server) containerImagePull(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImagePull(ctr.Params.Arguments["imageName"].(string))), nil
}
