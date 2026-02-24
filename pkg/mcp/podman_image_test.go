package mcp_test

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
	"github.com/manusa/podman-mcp-server/pkg/config"
)

// ImageSuite tests image tools using the mock Podman API server.
// These tests use the real podman CLI binary communicating with a mocked backend.
type ImageSuite struct {
	test.McpSuite
	containerFile string
}

// TestImageSuiteWithAllImplementations runs image tests with all implementations.
func TestImageSuiteWithAllImplementations(t *testing.T) {
	for _, impl := range test.AvailableImplementations() {
		t.Run(impl, func(t *testing.T) {
			suite.Run(t, &ImageSuite{
				McpSuite: test.McpSuite{Config: config.Config{PodmanImpl: impl}},
			})
		})
	}
}

func (s *ImageSuite) TestImageList() {
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
		text := toolResult.Content[0].(*mcp.TextContent).Text

		expectedHeaders := regexp.MustCompile(`(?m)^REPOSITORY\s+TAG\s+DIGEST\s+IMAGE ID\s+CREATED\s+SIZE\s*$`)
		s.Regexpf(expectedHeaders, text, "expected headers not found in output:\n%s", text)

		s.Contains(text, "nginx", "should contain nginx image")
		s.Contains(text, "redis", "should contain redis image")
	})

	s.Run("mock server received image list request", func() {
		s.True(s.MockServer.HasRequest("GET", "/libpod/images/json"))
	})
}

func (s *ImageSuite) TestImageListEmpty() {
	s.WithImageList([]test.ImageListResponse{})

	toolResult, err := s.CallTool("image_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty or headers-only output", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		// Some podman versions print headers even when empty, others don't
		// Just verify no image data is present
		s.NotContains(text, "nginx", "should not contain image data")
	})
}

func (s *ImageSuite) SetupTest() {
	s.McpSuite.SetupTest()
	// Create a temporary Containerfile for build tests
	s.containerFile = test.CreateTempFile(s.T(), "Containerfile", "FROM alpine:latest\nRUN echo 'test'\n")
}

func (s *ImageSuite) TestImagePull() {
	s.Run("image_pull(imageName=nil) returns error", func() {
		toolResult, err := s.CallTool("image_pull", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.Contains(text, "imageName", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("image_pull(imageName=nonexistent:latest) returns not found error", func() {
		s.WithError("POST", "/libpod/images/pull", "/images/create",
			404, "image not found: nonexistent:latest")

		toolResult, err := s.CallTool("image_pull", map[string]interface{}{
			"imageName": "nonexistent:latest",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("image_pull(imageName=nginx:latest) retries with docker.io prefix on short-name error", func() {
		// Setup: mock server returns error for short names, success for docker.io/ prefix
		handler := func(w http.ResponseWriter, r *http.Request) {
			reference := r.URL.Query().Get("reference")
			if strings.HasPrefix(reference, "docker.io/") {
				w.Header().Set("Content-Type", "application/json")
				test.WriteJSON(w, test.ImagePullResponse{
					ID:     "sha256:shortname123",
					Images: []string{"sha256:shortname123"},
					Status: "Download complete",
				})
				return
			}
			test.WriteError(w, http.StatusInternalServerError, "Error: short-name \""+reference+"\" did not resolve to an alias")
		}
		s.MockServer.HandleFunc("POST", "/libpod/images/pull", "/images/create", handler)

		toolResult, err := s.CallTool("image_pull", map[string]interface{}{
			"imageName": "nginx:latest",
		})

		s.Run("returns OK", func() {
			s.NoError(err)
			s.False(toolResult.IsError)
		})

		s.Run("returns success message with docker.io prefix", func() {
			text := toolResult.Content[0].(*mcp.TextContent).Text
			s.Contains(text, "docker.io/nginx", "should contain the docker.io prefixed image name")
			s.Contains(text, "pulled successfully", "should indicate success")
		})
	})

	s.Run("image_pull(imageName=docker.io/library/nginx:latest) pulls image", func() {
		s.WithImagePull("sha256:abc123def456")

		toolResult, err := s.CallTool("image_pull", map[string]interface{}{
			"imageName": "docker.io/library/nginx:latest",
		})

		s.Run("returns OK", func() {
			s.NoError(err)
			s.False(toolResult.IsError)
		})

		s.Run("returns success message", func() {
			text := toolResult.Content[0].(*mcp.TextContent).Text
			s.Contains(text, "nginx", "should contain image name")
			s.Contains(text, "pulled successfully", "should indicate success")
		})

		s.Run("mock server received pull request", func() {
			s.True(s.MockServer.HasRequest("POST", "/libpod/images/pull"))
		})
	})
}

func (s *ImageSuite) TestImagePush() {
	s.Run("image_push(imageName=nil) returns error", func() {
		toolResult, err := s.CallTool("image_push", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.Contains(text, "imageName", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("image_push(imageName=nonexistent:latest) returns not found error", func() {
		s.WithError("POST", "/libpod/images/{name}/push", "/images/{name}/push",
			404, "no such image: nonexistent:latest")

		toolResult, err := s.CallTool("image_push", map[string]interface{}{
			"imageName": "nonexistent:latest",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("image_push(imageName=example.com/org/image:tag) pushes image", func() {
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
			text := toolResult.Content[0].(*mcp.TextContent).Text
			s.Contains(text, "pushed successfully", "should indicate success")
		})

		s.Run("mock server received push request", func() {
			s.True(s.MockServer.HasRequest("POST", "/libpod/images/{name}/push"))
		})
	})
}

func (s *ImageSuite) TestImageRemove() {
	s.Run("image_remove(imageName=nil) returns error", func() {
		toolResult, err := s.CallTool("image_remove", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.Contains(text, "imageName", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("image_remove(imageName=nonexistent:latest) returns not found error", func() {
		s.WithError("DELETE", "/libpod/images/remove", "/images/remove",
			404, "no such image: nonexistent:latest")

		toolResult, err := s.CallTool("image_remove", map[string]interface{}{
			"imageName": "nonexistent:latest",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("image_remove(imageName=example.com/org/image:tag) removes image", func() {
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
			text := toolResult.Content[0].(*mcp.TextContent).Text
			s.NotEmpty(text, "should have output")
		})

		s.Run("mock server received remove request", func() {
			s.True(s.MockServer.HasRequest("DELETE", "/libpod/images/remove"))
		})
	})
}

func (s *ImageSuite) TestImageBuild() {
	s.Run("image_build(containerFile=nil) returns error", func() {
		toolResult, err := s.CallTool("image_build", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.Contains(text, "containerFile", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("image_build(containerFile=/nonexistent/Containerfile) returns not found error", func() {
		toolResult, err := s.CallTool("image_build", map[string]interface{}{
			"containerFile": "/nonexistent/Containerfile",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("image_build(containerFile=valid) builds image", func() {
		s.WithImageBuild("a1b2c3d4e5f6")

		_, _ = s.CallTool("image_build", map[string]interface{}{
			"containerFile": s.containerFile,
		})

		// Note: The mock server may not perfectly simulate the build streaming response,
		// so we focus on verifying the API was called correctly with the right parameters.

		s.Run("mock server received build request", func() {
			s.True(s.MockServer.HasRequest("POST", "/libpod/build"))
		})

		s.Run("build request includes dockerfile parameter", func() {
			req := s.PopLastCapturedRequest("POST", "/libpod/build")
			s.Require().NotNil(req, "build request should be captured")
			s.Contains(req.Query, "dockerfile=", "should have dockerfile query param")
		})
	})

	s.Run("image_build(imageName=example.com/org/image:tag) includes tag parameter", func() {
		s.WithImageBuild("b2c3d4e5f6a7")

		_, _ = s.CallTool("image_build", map[string]interface{}{
			"containerFile": s.containerFile,
			"imageName":     "example.com/org/image:tag",
		})

		s.Run("build request includes tag parameter", func() {
			req := s.PopLastCapturedRequest("POST", "/libpod/build")
			s.Require().NotNil(req, "build request should be captured")
			s.Contains(req.Query, "t=example.com", "should have tag query param")
		})
	})
}
