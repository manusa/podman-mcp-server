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
			mcp.WithDescription("Copies (pulls) a Docker or Podman container image from a registry onto the local machine storage"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to pull"), mcp.Required()),
		), s.containerImagePull},
		{mcp.NewTool("image_push",
			mcp.WithDescription("Pushes a Docker or Podman container image, manifest list or image index from local machine storage to a registry"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to push"), mcp.Required()),
		), s.containerImagePush},
		{mcp.NewTool("image_remove",
			mcp.WithDescription("Removes a Docker or Podman image from the local machine storage"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to remove"), mcp.Required()),
		), s.containerImageRemove},
	}
}

func (s *Server) containerImageList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImageList()), nil
}

func (s *Server) containerImagePull(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImagePull(ctr.Params.Arguments["imageName"].(string))), nil
}

func (s *Server) containerImagePush(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImagePush(ctr.Params.Arguments["imageName"].(string))), nil
}

func (s *Server) containerImageRemove(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImageRemove(ctr.Params.Arguments["imageName"].(string))), nil
}
