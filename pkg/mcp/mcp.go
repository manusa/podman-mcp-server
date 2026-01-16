package mcp

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/manusa/podman-mcp-server/pkg/podman"
	"github.com/manusa/podman-mcp-server/pkg/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and podman client.
type Server struct {
	server *mcp.Server
	podman podman.Podman
}

// NewServer creates a new MCP server with all tools registered.
func NewServer() (*Server, error) {
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
	if s.podman, err = podman.NewPodman(); err != nil {
		return nil, err
	}

	// Register all tools
	allTools := slices.Concat(
		initContainerTools(),
		initImageTools(),
		initNetworkTools(),
		initVolumeTools(),
	)

	for _, tool := range allTools {
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
