package mcp_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

type ContainerToolsSuite struct {
	test.McpSuite
}

func TestContainerTools(t *testing.T) {
	suite.Run(t, new(ContainerToolsSuite))
}

func (s *ContainerToolsSuite) TestContainerInspect() {
	toolResult, err := s.CallTool("container_inspect", map[string]interface{}{
		"name": "example-container",
	})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("inspects provided container", func() {
		s.Regexp("^podman inspect example-container", toolResult.Content[0].(mcp.TextContent).Text)
	})
}

func (s *ContainerToolsSuite) TestContainerList() {
	toolResult, err := s.CallTool("container_list", map[string]interface{}{})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("lists all containers", func() {
		s.Regexp("^podman container list -a", toolResult.Content[0].(mcp.TextContent).Text)
	})
}

func (s *ContainerToolsSuite) TestContainerLogs() {
	toolResult, err := s.CallTool("container_logs", map[string]interface{}{
		"name": "example-container",
	})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("retrieves logs from provided container", func() {
		s.Regexp("^podman logs example-container", toolResult.Content[0].(mcp.TextContent).Text)
	})
}

func (s *ContainerToolsSuite) TestContainerRemove() {
	toolResult, err := s.CallTool("container_remove", map[string]interface{}{
		"name": "example-container",
	})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("removes provided container", func() {
		s.Regexp("^podman container rm example-container", toolResult.Content[0].(mcp.TextContent).Text)
	})
}

func (s *ContainerToolsSuite) TestContainerRun() {
	s.Run("basic run", func() {
		toolResult, err := s.CallTool("container_run", map[string]interface{}{
			"imageName": "example.com/org/image:tag",
		})
		s.NoError(err)
		s.False(toolResult.IsError)
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Regexp("example.com/org/image:tag\n$", text)
		s.Contains(text, " -d ", "should run in detached mode")
		s.Contains(text, " --publish-all ", "should publish all exposed ports")
	})

	s.Run("with ports", func() {
		toolResult, err := s.CallTool("container_run", map[string]interface{}{
			"imageName": "example.com/org/image:tag",
			"ports": []interface{}{
				1337, // Invalid entry to test
				"8080:80",
				"8082:8082",
				"8443:443",
			},
		})
		s.NoError(err)
		s.False(toolResult.IsError)
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, " --publish=8080:80 ")
		s.Contains(text, " --publish=8082:8082 ")
		s.Contains(text, " --publish=8443:443 ")
	})

	s.Run("with environment", func() {
		toolResult, err := s.CallTool("container_run", map[string]interface{}{
			"imageName": "example.com/org/image:tag",
			"ports":     []interface{}{"8080:80"},
			"environment": []interface{}{
				"KEY=VALUE",
				"FOO=BAR",
			},
		})
		s.NoError(err)
		s.False(toolResult.IsError)
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, " --env KEY=VALUE ")
		s.Contains(text, " --env FOO=BAR ")
	})
}

func (s *ContainerToolsSuite) TestContainerStop() {
	toolResult, err := s.CallTool("container_stop", map[string]interface{}{
		"name": "example-container",
	})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("stops provided container", func() {
		s.Regexp("^podman container stop example-container", toolResult.Content[0].(mcp.TextContent).Text)
	})
}
