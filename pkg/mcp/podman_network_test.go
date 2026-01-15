package mcp_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

type NetworkToolsSuite struct {
	test.McpSuite
}

func TestNetworkTools(t *testing.T) {
	suite.Run(t, new(NetworkToolsSuite))
}

func (s *NetworkToolsSuite) TestNetworkList() {
	toolResult, err := s.CallTool("network_list", map[string]interface{}{})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("lists all available networks", func() {
		s.Regexp("^podman network ls", toolResult.Content[0].(mcp.TextContent).Text)
	})
}
