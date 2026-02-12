package mcp

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/manusa/podman-mcp-server/pkg/api"
	"github.com/manusa/podman-mcp-server/pkg/podman"
	"github.com/manusa/podman-mcp-server/pkg/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and podman client.
type Server struct {
	server *mcp.Server
	podman podman.Podman
}

// serverConfig holds the configuration for the MCP server.
type serverConfig struct {
	podmanImpl string
}

// ServerOption configures the MCP server.
type ServerOption func(*serverConfig)

// WithPodmanImpl sets the Podman implementation to use.
// If not specified, auto-detects the best available implementation.
// Valid values: "cli" (default), "api" (future)
func WithPodmanImpl(impl string) ServerOption {
	return func(c *serverConfig) {
		c.podmanImpl = impl
	}
}

// NewServer creates a new MCP server with all tools registered.
// Use functional options to configure the server:
//
//	server, err := mcp.NewServer(mcp.WithPodmanImpl("cli"))
func NewServer(opts ...ServerOption) (*Server, error) {
	cfg := &serverConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

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
	if s.podman, err = podman.NewPodman(cfg.podmanImpl); err != nil {
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
	return slices.Concat(
		initContainerTools(),
		initImageTools(),
		initNetworkTools(),
		initVolumeTools(),
	)
}

// ServeStreamableHTTP returns an HTTP handler for Streamable HTTP transport.
func (s *Server) ServeStreamableHTTP() *mcp.StreamableHTTPHandler {
	return mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return s.server
	}, nil)
}
