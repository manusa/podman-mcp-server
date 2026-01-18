package mcp_test

import (
	"regexp"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

// ImageToolsSuite tests image tools using the mock Podman API server.
// These tests use the real podman CLI binary communicating with a mocked backend.
type ImageToolsSuite struct {
	test.MockServerMcpSuite
}

func TestImageTools(t *testing.T) {
	suite.Run(t, new(ImageToolsSuite))
}

func (s *ImageToolsSuite) TestImageList() {
	s.WithImageList([]test.ImageListResponse{
		{
			ID:          "sha256:abc123def456",
			RepoTags:    []string{"docker.io/library/nginx:latest"},
			RepoDigests: []string{"docker.io/library/nginx@sha256:abc123"},
			Created:     1704067200,
			Size:        142000000,
			VirtualSize: 142000000,
			Labels:      map[string]string{"maintainer": "nginx"},
		},
		{
			ID:          "sha256:xyz789ghi012",
			RepoTags:    []string{"docker.io/library/redis:alpine"},
			RepoDigests: []string{"docker.io/library/redis@sha256:xyz789"},
			Created:     1704067200,
			Size:        37000000,
			VirtualSize: 37000000,
		},
	})

	toolResult, err := s.CallTool("image_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns image data with expected format", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text

		expectedHeaders := regexp.MustCompile(`(?m)^REPOSITORY\s+TAG\s+DIGEST\s+IMAGE ID\s+CREATED\s+SIZE\s*$`)
		s.Regexpf(expectedHeaders, text, "expected headers not found in output:\n%s", text)

		s.Contains(text, "nginx", "should contain nginx image")
		s.Contains(text, "redis", "should contain redis image")
	})

	s.Run("mock server received image list request", func() {
		s.True(s.MockServer.HasRequest("GET", "/libpod/images/json"))
	})
}

func (s *ImageToolsSuite) TestImageListEmpty() {
	s.WithImageList([]test.ImageListResponse{})

	toolResult, err := s.CallTool("image_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty or headers-only output", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		// Some podman versions print headers even when empty, others don't
		// Just verify no image data is present
		s.NotContains(text, "nginx", "should not contain image data")
	})
}

func (s *ImageToolsSuite) TestImagePull() {
	s.WithImagePull("sha256:abc123def456")

	toolResult, err := s.CallTool("image_pull", map[string]interface{}{
		"imageName": "docker.io/library/nginx:latest",
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns success message", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "nginx", "should contain image name")
		s.Contains(text, "pulled successfully", "should indicate success")
	})

	s.Run("mock server received pull request", func() {
		s.True(s.MockServer.HasRequest("POST", "/libpod/images/pull"))
	})
}

func (s *ImageToolsSuite) TestImagePush() {
	// First need an image to exist (podman checks it before pushing)
	s.WithImageList([]test.ImageListResponse{
		{
			ID:       "sha256:abc123def456",
			RepoTags: []string{"example.com/org/image:tag"},
			Created:  1704067200,
			Size:     142000000,
		},
	})
	s.WithImagePush()

	toolResult, err := s.CallTool("image_push", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns success message", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "pushed successfully", "should indicate success")
	})

	s.Run("mock server received push request", func() {
		s.True(s.MockServer.HasRequest("POST", "/libpod/images/{name}/push"))
	})
}

func (s *ImageToolsSuite) TestImageRemove() {
	// Podman looks up the image before removing
	s.WithImageList([]test.ImageListResponse{
		{
			ID:       "sha256:abc123def456",
			RepoTags: []string{"example.com/org/image:tag"},
			Created:  1704067200,
			Size:     142000000,
		},
	})
	s.WithImageRemove("sha256:abc123def456")

	toolResult, err := s.CallTool("image_remove", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns success response", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.NotEmpty(text, "should have output")
	})

	s.Run("mock server received remove request", func() {
		s.True(s.MockServer.HasRequest("DELETE", "/libpod/images/remove"))
	})
}

func (s *ImageToolsSuite) TestImagePullNotFound() {
	s.WithError("POST", "/libpod/images/pull", "/images/create",
		404, "image not found: nonexistent:latest")

	toolResult, err := s.CallTool("image_pull", map[string]interface{}{
		"imageName": "nonexistent:latest",
	})

	s.Run("returns error", func() {
		s.NoError(err) // MCP call succeeds
		s.True(toolResult.IsError, "tool result should indicate an error")
	})

	s.Run("error message indicates failure", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})
}

// ImageBuildSuite tests the image_build tool using a fake podman binary.
// This suite validates that CLI arguments are correctly constructed.
type ImageBuildSuite struct {
	test.McpSuite
}

func TestImageBuild(t *testing.T) {
	suite.Run(t, new(ImageBuildSuite))
}

func (s *ImageBuildSuite) TestImageBuildBasic() {
	toolResult, err := s.CallTool("image_build", map[string]interface{}{
		"containerFile": "/tmp/Containerfile",
	})
	s.NoError(err)
	s.False(toolResult.IsError)
	s.Regexp("^podman build -f /tmp/Containerfile", toolResult.Content[0].(mcp.TextContent).Text)
}

func (s *ImageBuildSuite) TestImageBuildWithImageName() {
	toolResult, err := s.CallTool("image_build", map[string]interface{}{
		"containerFile": "/tmp/Containerfile",
		"imageName":     "example.com/org/image:tag",
	})
	s.NoError(err)
	s.False(toolResult.IsError)
	s.Regexp("^podman build -t example.com/org/image:tag -f /tmp/Containerfile", toolResult.Content[0].(mcp.TextContent).Text)
}
