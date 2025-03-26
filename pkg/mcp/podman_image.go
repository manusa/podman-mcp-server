package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initPodmanImage() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("image_build",
			mcp.WithDescription("Build a Docker or Podman image from a Dockerfile, Podmanfile, or Containerfile"),
			mcp.WithString("containerFile", mcp.Description("The absolute path to the Dockerfile, Podmanfile, or Containerfile to build the image from"), mcp.Required()),
			mcp.WithString("imageName", mcp.Description("Specifies the name which is assigned to the resulting image if the build process completes successfully (--tag, -t)")),
		), s.imageBuild},
		{mcp.NewTool("image_list",
			mcp.WithDescription("List the Docker or Podman images on the local machine"),
		), s.imageList},
		{mcp.NewTool("image_pull",
			mcp.WithDescription("Copies (pulls) a Docker or Podman container image from a registry onto the local machine storage"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to pull"), mcp.Required()),
		), s.imagePull},
		{mcp.NewTool("image_push",
			mcp.WithDescription("Pushes a Docker or Podman container image, manifest list or image index from local machine storage to a registry"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to push"), mcp.Required()),
		), s.imagePush},
		{mcp.NewTool("image_remove",
			mcp.WithDescription("Removes a Docker or Podman image from the local machine storage"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to remove"), mcp.Required()),
		), s.imageRemove},
	}
}

func (s *Server) imageBuild(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imageName := ctr.Params.Arguments["imageName"]
	if _, ok := imageName.(string); !ok {
		imageName = ""
	}
	return NewTextResult(s.podman.ImageBuild(ctr.Params.Arguments["containerFile"].(string), imageName.(string))), nil
}

func (s *Server) imageList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImageList()), nil
}

func (s *Server) imagePull(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImagePull(ctr.Params.Arguments["imageName"].(string))), nil
}

func (s *Server) imagePush(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImagePush(ctr.Params.Arguments["imageName"].(string))), nil
}

func (s *Server) imageRemove(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ImageRemove(ctr.Params.Arguments["imageName"].(string))), nil
}
