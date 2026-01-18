package mcp_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

// ContainerSuite tests container tools using the mock Podman API server.
// These tests use the real podman CLI binary communicating with a mocked backend.
type ContainerSuite struct {
	test.McpSuite
}

func TestContainerSuite(t *testing.T) {
	suite.Run(t, new(ContainerSuite))
}

func (s *ContainerSuite) TestContainerList() {
	s.WithContainerList([]test.ContainerListResponse{
		{
			ID:        "abc123def456",
			Names:     []string{"test-container-1"},
			Image:     "docker.io/library/nginx:latest",
			ImageID:   "sha256:abc123",
			State:     "running",
			Status:    "Up 2 hours",
			Created:   "2024-01-01T00:00:00Z", // RFC3339 time string for libpod
			Command:   []string{"/bin/sh"},
			StartedAt: 1704067200,
		},
		{
			ID:        "xyz789ghi012",
			Names:     []string{"test-container-2"},
			Image:     "docker.io/library/redis:alpine",
			ImageID:   "sha256:xyz789",
			State:     "exited",
			Status:    "Exited (0) 1 hour ago",
			Created:   "2024-01-01T00:00:00Z",
			Command:   []string{"/bin/sh"},
			StartedAt: 1704067200,
		},
	})

	toolResult, err := s.CallTool("container_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns container data with expected format", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text

		expectedHeaders := regexp.MustCompile(`(?m)^CONTAINER ID\s+IMAGE\s+COMMAND\s+CREATED\s+STATUS\s+PORTS\s+NAMES\s*$`)
		s.Regexpf(expectedHeaders, text, "expected headers not found in output:\n%s", text)

		expectedRows := []string{
			`abc123def456\s+.*nginx.*\s+.*\s+.*\s+.*\s+.*\s+test-container-1`,
			`xyz789ghi012\s+.*redis.*\s+.*\s+.*\s+.*\s+.*\s+test-container-2`,
		}
		for _, row := range expectedRows {
			s.Regexpf(row, text, "expected row '%s' not found in output:\n%s", row, text)
		}
	})

	s.Run("mock server received container list request", func() {
		s.True(s.MockServer.HasRequest("GET", "/libpod/containers/json"))
	})
}

func (s *ContainerSuite) TestContainerListEmpty() {
	s.WithContainerList([]test.ContainerListResponse{})

	toolResult, err := s.CallTool("container_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty or headers-only output", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		// Some podman versions print headers even when empty, others don't
		// Just verify no container data is present
		s.NotContains(text, "test-container", "should not contain container data")
	})
}

func (s *ContainerSuite) TestContainerInspect() {
	s.WithContainerInspect(test.ContainerInspectResponse{
		ID:        "abc123def456",
		Name:      "/test-container",
		Image:     "sha256:abc123",
		ImageName: "docker.io/library/nginx:latest",
		Created:   "2024-01-01T00:00:00Z",
		State: &test.ContainerState{
			Status:    "running",
			Running:   true,
			StartedAt: "2024-01-01T00:00:00Z",
		},
		Config: &test.ContainerConfig{
			Image:    "docker.io/library/nginx:latest",
			Hostname: "abc123def456",
			Env:      []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		},
	})

	toolResult, err := s.CallTool("container_inspect", map[string]interface{}{
		"name": "test-container",
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns container details with expected format", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text

		// Inspect returns JSON format with container details
		s.Contains(text, "abc123def456", "should contain container ID")
		s.Contains(text, "nginx", "should contain image name")
		s.Contains(text, "running", "should contain container state")
	})

	s.Run("mock server received inspect request", func() {
		s.True(s.MockServer.HasRequest("GET", "/libpod/containers/{id}/json"))
	})
}

func (s *ContainerSuite) TestContainerInspectNotFound() {
	s.WithError("GET", "/libpod/containers/{id}/json", "/containers/{id}/json",
		404, "no such container: nonexistent")

	toolResult, err := s.CallTool("container_inspect", map[string]interface{}{
		"name": "nonexistent",
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

func (s *ContainerSuite) TestContainerLogs() {
	// Podman CLI first inspects the container before fetching logs
	s.WithContainerInspect(test.ContainerInspectResponse{
		ID:        "abc123def456",
		Name:      "/test-container",
		Image:     "sha256:abc123",
		ImageName: "docker.io/library/nginx:latest",
		Created:   "2024-01-01T00:00:00Z",
		State: &test.ContainerState{
			Status:    "running",
			Running:   true,
			StartedAt: "2024-01-01T00:00:00Z",
		},
	})
	expectedLogs := "2024-01-01T00:00:00Z Starting nginx...\n2024-01-01T00:00:01Z nginx started successfully\n"
	s.WithContainerLogs(expectedLogs)

	toolResult, err := s.CallTool("container_logs", map[string]interface{}{
		"name": "test-container",
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns log content with expected format", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text

		// Verify log lines are present and in correct order
		s.Contains(text, "Starting nginx", "should contain first log line")
		s.Contains(text, "nginx started successfully", "should contain second log line")
		s.Less(
			strings.Index(text, "Starting nginx"),
			strings.Index(text, "nginx started successfully"),
			"log lines should appear in chronological order",
		)
	})

	s.Run("mock server received logs request", func() {
		s.True(s.MockServer.HasRequest("GET", "/libpod/containers/{id}/logs"))
	})
}

func (s *ContainerSuite) TestContainerStop() {
	// Podman first looks up container by name, then inspects, then stops
	s.WithContainerList([]test.ContainerListResponse{
		{
			ID:        "abc123def456",
			Names:     []string{"test-container"},
			Image:     "docker.io/library/nginx:latest",
			State:     "running",
			Created:   "2024-01-01T00:00:00Z",
			Command:   []string{"/bin/sh"},
			StartedAt: 1704067200,
		},
	})
	s.WithContainerInspect(test.ContainerInspectResponse{
		ID:      "abc123def456",
		Name:    "/test-container",
		Created: "2024-01-01T00:00:00Z",
		State:   &test.ContainerState{Status: "running", Running: true},
	})
	s.WithContainerStop()

	toolResult, err := s.CallTool("container_stop", map[string]interface{}{
		"name": "test-container",
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns success response with container name", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "test-container", "should contain the stopped container name")
	})

	s.Run("mock server received stop request", func() {
		s.True(s.MockServer.HasRequest("POST", "/libpod/containers/{id}/stop"))
	})
}

func (s *ContainerSuite) TestContainerRemove() {
	// Podman first looks up container by name, then inspects, then removes
	s.WithContainerList([]test.ContainerListResponse{
		{
			ID:        "abc123def456",
			Names:     []string{"test-container"},
			Image:     "docker.io/library/nginx:latest",
			State:     "exited",
			Created:   "2024-01-01T00:00:00Z",
			Command:   []string{"/bin/sh"},
			StartedAt: 1704067200,
		},
	})
	s.WithContainerInspect(test.ContainerInspectResponse{
		ID:      "abc123def456",
		Name:    "/test-container",
		Created: "2024-01-01T00:00:00Z",
		State:   &test.ContainerState{Status: "exited", Running: false},
	})
	s.WithContainerRemove()

	toolResult, err := s.CallTool("container_remove", map[string]interface{}{
		"name": "test-container",
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns success response with container name", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "test-container", "should contain the removed container name")
	})

	s.Run("mock server received remove request", func() {
		s.True(s.MockServer.HasRequest("DELETE", "/libpod/containers/{id}"))
	})
}

func (s *ContainerSuite) TestContainerRunBasic() {
	s.WithContainerRun("container123")

	toolResult, err := s.CallTool("container_run", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns container ID", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "container123")
	})

	s.Run("mock server received create request", func() {
		s.True(s.MockServer.HasRequest("POST", "/libpod/containers/create"))
	})

	s.Run("mock server received start request", func() {
		s.True(s.MockServer.HasRequest("POST", "/libpod/containers/{id}/start"))
	})
}

func (s *ContainerSuite) TestContainerRunWithPorts() {
	s.WithContainerRun("container-with-ports")

	toolResult, err := s.CallTool("container_run", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
		"ports": []interface{}{
			1337, // Invalid entry - should be ignored
			"8080:80",
			"8082:8082",
			"8443:443",
		},
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("create request includes port mappings", func() {
		req := s.GetCapturedRequest("POST", "/libpod/containers/create")
		s.Require().NotNil(req, "create request should be captured")
		// Verify port mappings are in the request body
		s.Contains(req.Body, `"host_port":8080`, "should have port 8080 mapping")
		s.Contains(req.Body, `"host_port":8082`, "should have port 8082 mapping")
		s.Contains(req.Body, `"host_port":8443`, "should have port 8443 mapping")
	})
}

func (s *ContainerSuite) TestContainerRunWithEnvironment() {
	s.WithContainerRun("container-with-env")

	toolResult, err := s.CallTool("container_run", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
		"ports":     []interface{}{"8080:80"},
		"environment": []interface{}{
			"KEY=VALUE",
			"FOO=BAR",
		},
	})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("create request includes environment variables", func() {
		req := s.GetCapturedRequest("POST", "/libpod/containers/create")
		s.Require().NotNil(req, "create request should be captured")
		// Verify environment variables are in the request body
		s.Contains(req.Body, `"KEY":"VALUE"`, "should have KEY=VALUE env var")
		s.Contains(req.Body, `"FOO":"BAR"`, "should have FOO=BAR env var")
	})
}
