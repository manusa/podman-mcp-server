package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

type McpServerSuite struct {
	test.McpSuite
}

func TestMcpServer(t *testing.T) {
	suite.Run(t, new(McpServerSuite))
}

func (s *McpServerSuite) TestListTools() {
	expectedTools := []string{
		"container_inspect",
		"container_list",
		"container_logs",
		"container_remove",
		"container_run",
		"container_stop",
		"image_build",
		"image_list",
		"image_pull",
		"image_push",
		"image_remove",
		"network_list",
		"volume_list",
	}

	tools, err := s.ListTools()
	s.Require().NoError(err)

	nameSet := make(map[string]bool)
	for _, tool := range tools.Tools {
		nameSet[tool.Name] = true
	}

	for _, name := range expectedTools {
		s.Run("has "+name+" tool", func() {
			s.True(nameSet[name], "tool %s not found", name)
		})
	}
}
