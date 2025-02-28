package mcp

import (
	"github.com/manusa/podman-mcp-server/pkg/version"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	server *server.MCPServer
}

func NewSever() (*Server, error) {
	s := &Server{
		server: server.NewMCPServer(
			version.BinaryName,
			version.Version,
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithLogging(),
		),
	}
	//s.server.AddTools()
	return s, nil
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}
