package mcp

import (
	"context"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initPodmanContainer() []server.ServerTool {
	return []server.ServerTool{
		{mcp.NewTool("container_list",
			mcp.WithDescription("Prints out information about the running Docker or Podman containers"),
		), s.containerList},
		{mcp.NewTool("container_run",
			mcp.WithDescription("Runs a Docker or Podman container with the specified image name"),
			mcp.WithString("imageName", mcp.Description("Docker or Podman container image name to pull"), mcp.Required()),
		), s.containerRun},
		{mcp.NewTool("container_stop",
			mcp.WithDescription("Stops a Docker or Podman running container with the specified container ID or name"),
			mcp.WithString("name", mcp.Description("Docker or Podman container ID or name to stop"), mcp.Required()),
		), s.containerStop},
	}
}

func (s *Server) containerList(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ContainerList()), nil
}

func (s *Server) containerRun(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ContainerRun(ctr.Params.Arguments["imageName"].(string))), nil
}

func (s *Server) containerStop(_ context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return NewTextResult(s.podman.ContainerStop(ctr.Params.Arguments["name"].(string))), nil
}
