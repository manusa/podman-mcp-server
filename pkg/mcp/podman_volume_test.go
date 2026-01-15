package mcp_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

type VolumeToolsSuite struct {
	test.McpSuite
}

func TestVolumeTools(t *testing.T) {
	suite.Run(t, new(VolumeToolsSuite))
}

func (s *VolumeToolsSuite) TestVolumeList() {
	toolResult, err := s.CallTool("volume_list", map[string]interface{}{})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("lists all available volumes", func() {
		s.Regexp("^podman volume ls", toolResult.Content[0].(mcp.TextContent).Text)
	})
}
