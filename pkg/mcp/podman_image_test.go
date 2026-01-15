package mcp_test

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

type ImageToolsSuite struct {
	test.McpSuite
}

func TestImageTools(t *testing.T) {
	suite.Run(t, new(ImageToolsSuite))
}

func (s *ImageToolsSuite) TestImageBuild() {
	s.Run("basic build", func() {
		toolResult, err := s.CallTool("image_build", map[string]interface{}{
			"containerFile": "/tmp/Containerfile",
		})
		s.NoError(err)
		s.False(toolResult.IsError)
		s.Regexp("^podman build -f /tmp/Containerfile", toolResult.Content[0].(mcp.TextContent).Text)
	})

	s.Run("with imageName", func() {
		toolResult, err := s.CallTool("image_build", map[string]interface{}{
			"containerFile": "/tmp/Containerfile",
			"imageName":     "example.com/org/image:tag",
		})
		s.NoError(err)
		s.False(toolResult.IsError)
		s.Regexp("^podman build -t example.com/org/image:tag -f /tmp/Containerfile", toolResult.Content[0].(mcp.TextContent).Text)
	})
}

func (s *ImageToolsSuite) TestImageList() {
	s.WithPodmanOutput(
		"REPOSITORY\tTAG\tDIGEST\tIMAGE ID\tCREATED\tSIZE",
		"docker.io/marcnuri/chuck-norris\nlatest\nsha256:1337\nb8f22a2b8410\n1 day ago\n37 MB",
	)
	toolResult, err := s.CallTool("image_list", map[string]interface{}{})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("lists images with digests", func() {
		s.Regexp("^podman images --digests", toolResult.Content[0].(mcp.TextContent).Text)
	})
}

func (s *ImageToolsSuite) TestImagePull() {
	toolResult, err := s.CallTool("image_pull", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
	})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("pulls specified image", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Regexp("^podman image pull example.com/org/image:tag", text)
		s.Regexp("example.com/org/image:tag pulled successfully$", text)
	})
}

func (s *ImageToolsSuite) TestImagePush() {
	toolResult, err := s.CallTool("image_push", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
	})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("pushes specified image", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Regexp("^podman image push example.com/org/image:tag", text)
		s.Regexp("example.com/org/image:tag pushed successfully$", text)
	})
}

func (s *ImageToolsSuite) TestImageRemove() {
	toolResult, err := s.CallTool("image_remove", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
	})
	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})
	s.Run("removes specified image", func() {
		s.Regexp("^podman image rm example.com/org/image:tag", toolResult.Content[0].(mcp.TextContent).Text)
	})
}
