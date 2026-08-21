package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/manusa/podman-mcp-server/pkg/api"
	"github.com/manusa/podman-mcp-server/pkg/config"
	"github.com/manusa/podman-mcp-server/pkg/podman"
	"github.com/manusa/podman-mcp-server/pkg/toolsets"
	"github.com/manusa/podman-mcp-server/pkg/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and podman client.
type Server struct {
	server *mcp.Server
	podman podman.Podman
}

// NewServer creates a new MCP server with all tools registered.
func NewServer(cfg config.Config) (*Server, error) {
	s := &Server{
		server: mcp.NewServer(
			&mcp.Implementation{
				Name:    version.BinaryName,
				Version: version.Version,
			},
			&mcp.ServerOptions{
				Capabilities: &mcp.ServerCapabilities{
					Tools:   &mcp.ToolCapabilities{},
					Logging: &mcp.LoggingCapabilities{},
				},
			},
		),
	}

	var err error
	if s.podman, err = podman.NewPodman(cfg); err != nil {
		return nil, err
	}

	// Register all tools
	for _, tool := range AllTools() {
		goSdkTool, handler, err := ServerToolToGoSdkTool(s.podman, tool)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool %s: %w", tool.Tool.Name, err)
		}
		s.server.AddTool(goSdkTool, handler)
	}

	return s, nil
}

// ServeStdio starts the server using STDIO transport.
func (s *Server) ServeStdio(ctx context.Context) error {
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

// ServeSse returns an HTTP handler for SSE transport.
func (s *Server) ServeSse() *mcp.SSEHandler {
	return mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return s.server
	}, nil)
}

// AllTools returns all registered tools for documentation purposes.
func AllTools() []api.ServerTool {
	var all []api.ServerTool
	for _, ts := range toolsets.Toolsets() {
		all = append(all, ts.GetTools()...)
	}
	return all
}

// ServeStreamableHTTP returns an HTTP handler for Streamable HTTP transport.
func (s *Server) ServeStreamableHTTP() *mcp.StreamableHTTPHandler {
	return mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return s.server
	}, nil)
}
