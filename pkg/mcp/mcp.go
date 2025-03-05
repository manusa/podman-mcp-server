package mcp

import (
	"github.com/manusa/podman-mcp-server/pkg/podman"
	"github.com/manusa/podman-mcp-server/pkg/version"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"slices"
	"strconv"
)

type Server struct {
	server *server.MCPServer
	podman podman.Podman
}

func NewSever() (*Server, error) {
	s := &Server{
		server: server.NewMCPServer(
			version.BinaryName,
			version.Version,
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithToolCapabilities(true),
			server.WithLogging(),
		),
	}
	var err error
	if s.podman, err = podman.NewPodman(); err != nil {
		return nil, err
	}
	s.server.AddTools(slices.Concat(
		s.initPodmanContainer(),
		s.initPodmanImage(),
	)...)
	return s, nil
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}

func (s *Server) ServeSse(port int) *server.SSEServer {
	return server.NewSSEServer(s.server, "http://localhost:"+strconv.Itoa(port))
}

func NewTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: err.Error(),
				},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
	}
}
