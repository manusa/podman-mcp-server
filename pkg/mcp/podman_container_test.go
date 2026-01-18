package mcp_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

// ContainerToolsSuite tests container tools using the mock Podman API server.
// These tests use the real podman CLI binary communicating with a mocked backend.
type ContainerToolsSuite struct {
	test.MockServerMcpSuite
}

func TestContainerTools(t *testing.T) {
	suite.Run(t, new(ContainerToolsSuite))
}

func (s *ContainerToolsSuite) TestContainerList() {
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

func (s *ContainerToolsSuite) TestContainerInspect() {
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

func (s *ContainerToolsSuite) TestContainerLogs() {
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

func (s *ContainerToolsSuite) TestContainerStop() {
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

func (s *ContainerToolsSuite) TestContainerRemove() {
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

func (s *ContainerToolsSuite) TestContainerListEmpty() {
	s.WithContainerList([]test.ContainerListResponse{})

	toolResult, err := s.CallTool("container_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns headers only with no data rows", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text

		// Should have headers
		s.Contains(text, "CONTAINER ID", "should contain header")
		s.Contains(text, "IMAGE", "should contain header")
		s.Contains(text, "NAMES", "should contain header")

		// Should not have any container data rows (only header line)
		lines := strings.Split(strings.TrimSpace(text), "\n")
		s.Equal(1, len(lines), "expected only header line in output:\n%s", text)
	})
}

func (s *ContainerToolsSuite) TestContainerInspectNotFound() {
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

// ContainerRunSuite tests the container_run tool using a fake podman binary.
// This suite validates that CLI arguments are correctly constructed.
type ContainerRunSuite struct {
	test.McpSuite
}

func TestContainerRun(t *testing.T) {
	suite.Run(t, new(ContainerRunSuite))
}

func (s *ContainerRunSuite) TestContainerRunBasic() {
	toolResult, err := s.CallTool("container_run", map[string]interface{}{
		"imageName": "example.com/org/image:tag",
	})
	s.NoError(err)
	s.False(toolResult.IsError)
	text := toolResult.Content[0].(mcp.TextContent).Text
	s.Regexp("example.com/org/image:tag\n$", text)
	s.Contains(text, " -d ", "should run in detached mode")
	s.Contains(text, " --publish-all ", "should publish all exposed ports")
}

func (s *ContainerRunSuite) TestContainerRunWithPorts() {
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
}

func (s *ContainerRunSuite) TestContainerRunWithEnvironment() {
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
}
