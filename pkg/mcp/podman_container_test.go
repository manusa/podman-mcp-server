package mcp_test

import (
	"encoding/json"
	"net/http"
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

		var containers []test.ContainerListResponse
		s.Require().NoError(json.Unmarshal([]byte(text), &containers))

		s.Require().Len(containers, 2)

		containersByID := make(map[string]test.ContainerListResponse)
		for _, cont := range containers {
			containersByID[cont.ID] = cont
		}

		s.Contains(containersByID, "abc123def456", "should contain test-container-1 container")
		s.Contains(containersByID, "xyz789ghi012", "should contain test-container-2 container")

		nginxContainer := containersByID["abc123def456"]
		s.Contains(nginxContainer.Image, "nginx", "should contain nginx image")
		s.Require().NotEmpty(nginxContainer.Names, "should have container names")
		s.Contains(nginxContainer.Names[0], "test-container-1", "should have correct name")
		s.Equal("running", nginxContainer.State, "should be running")

		redisContainer := containersByID["xyz789ghi012"]
		s.Contains(redisContainer.Image, "redis", "should contain redis image")
		s.Require().NotEmpty(redisContainer.Names, "should have container names")
		s.Contains(redisContainer.Names[0], "test-container-2", "should have correct name")
		s.Equal("exited", redisContainer.State, "should be exited")
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
	s.Run("container_inspect(name=nil) returns error", func() {
		toolResult, err := s.CallTool("container_inspect", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "name", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("container_inspect(name=123) returns error for non-string parameter", func() {
		toolResult, err := s.CallTool("container_inspect", map[string]interface{}{
			"name": 123, // Pass an integer instead of a string
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "name", "error should mention the parameter")
		s.Contains(text, "must be a string", "error should indicate parameter must be a string")
	})

	s.Run("container_inspect(name=nonexistent) returns not found error", func() {
		s.WithError("GET", "/libpod/containers/{id}/json", "/containers/{id}/json",
			404, "no such container: nonexistent")

		toolResult, err := s.CallTool("container_inspect", map[string]interface{}{
			"name": "nonexistent",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("container_inspect(name=test-container) returns container details", func() {
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
			s.Contains(text, "abc123def456", "should contain container ID")
			s.Contains(text, "nginx", "should contain image name")
			s.Contains(text, "running", "should contain container state")
		})

		s.Run("mock server received inspect request", func() {
			s.True(s.MockServer.HasRequest("GET", "/libpod/containers/{id}/json"))
		})
	})
}

func (s *ContainerSuite) TestContainerLogs() {
	s.Run("container_logs(name=nil) returns error", func() {
		toolResult, err := s.CallTool("container_logs", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "name", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("container_logs(name=nonexistent) returns not found error", func() {
		s.WithError("GET", "/libpod/containers/{id}/json", "/containers/{id}/json",
			404, "no such container: nonexistent")

		toolResult, err := s.CallTool("container_logs", map[string]interface{}{
			"name": "nonexistent",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("container_logs(name=test-container) returns logs", func() {
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
	})
}

func (s *ContainerSuite) TestContainerStop() {
	s.Run("container_stop(name=nil) returns error", func() {
		toolResult, err := s.CallTool("container_stop", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "name", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("container_stop(name=nonexistent) returns not found error", func() {
		s.WithError("GET", "/libpod/containers/{id}/json", "/containers/{id}/json",
			404, "no such container: nonexistent")

		toolResult, err := s.CallTool("container_stop", map[string]interface{}{
			"name": "nonexistent",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("container_stop(name=test-container) stops container", func() {
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
	})
}

func (s *ContainerSuite) TestContainerRemove() {
	s.Run("container_remove(name=nil) returns error", func() {
		toolResult, err := s.CallTool("container_remove", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "name", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("container_remove(name=nonexistent) returns not found error", func() {
		s.WithError("GET", "/libpod/containers/{id}/json", "/containers/{id}/json",
			404, "no such container: nonexistent")

		toolResult, err := s.CallTool("container_remove", map[string]interface{}{
			"name": "nonexistent",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("container_remove(name=test-container) removes container", func() {
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
	})
}

func (s *ContainerSuite) TestContainerRun() {
	s.Run("container_run(imageName=nil) returns error", func() {
		toolResult, err := s.CallTool("container_run", map[string]interface{}{})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.Contains(text, "imageName", "error should mention the missing parameter")
		s.Contains(text, "required", "error should indicate parameter is required")
	})

	s.Run("container_run(imageName=nonexistent:latest) returns image not found error", func() {
		s.WithError("POST", "/libpod/containers/create", "/containers/create",
			404, "no such image: nonexistent:latest")

		toolResult, err := s.CallTool("container_run", map[string]interface{}{
			"imageName": "nonexistent:latest",
		})
		s.NoError(err)
		s.True(toolResult.IsError, "tool result should indicate an error")
		text := toolResult.Content[0].(mcp.TextContent).Text
		s.NotEmpty(text, "error message should not be empty")
	})

	s.Run("container_run(environment=string) ignores non-array environment parameter", func() {
		s.WithContainerRun("container-with-invalid-env")

		toolResult, err := s.CallTool("container_run", map[string]interface{}{
			"imageName":   "example.com/org/image:tag",
			"environment": "FOO=BAR", // Pass a string instead of an array
		})

		s.Run("returns OK", func() {
			s.NoError(err)
			s.False(toolResult.IsError)
		})

		s.Run("create request does not include environment variables", func() {
			req := s.PopLastCapturedRequest("POST", "/libpod/containers/create")
			s.Require().NotNil(req, "create request should be captured")
			s.NotContains(req.Body, `"FOO"`, "should not have env var from invalid parameter")
		})
	})

	s.Run("container_run(imageName=nginx:latest) retries with docker.io prefix on short-name error", func() {
		// Setup: mock server returns error for short names, success for docker.io/ prefix
		pullHandler := func(w http.ResponseWriter, r *http.Request) {
			reference := r.URL.Query().Get("reference")
			if strings.HasPrefix(reference, "docker.io/") {
				test.WriteJSON(w, test.ImagePullResponse{
					ID:     "sha256:abc123def456",
					Status: "Already exists",
				})
				return
			}
			test.WriteError(w, http.StatusInternalServerError, "Error: short-name \""+reference+"\" did not resolve to an alias")
		}
		s.MockServer.HandleFunc("POST", "/libpod/images/pull", "/images/create", pullHandler)

		imageInspectHandler := func(w http.ResponseWriter, _ *http.Request) {
			test.WriteJSON(w, map[string]any{
				"Id":       "sha256:abc123def456",
				"RepoTags": []string{"docker.io/library/nginx:latest"},
				"Config": map[string]any{
					"ExposedPorts": map[string]any{"80/tcp": map[string]any{}},
				},
			})
		}
		s.MockServer.Handle("GET", "/libpod/images/{name}/json", imageInspectHandler)

		createHandler := func(w http.ResponseWriter, _ *http.Request) {
			test.WriteJSON(w, map[string]string{"Id": "container-shortname"})
		}
		s.MockServer.HandleFunc("POST", "/libpod/containers/create", "/containers/create", createHandler)

		startHandler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}
		s.MockServer.Handle("POST", "/libpod/containers/{id}/start", startHandler)

		toolResult, err := s.CallTool("container_run", map[string]interface{}{
			"imageName": "nginx:latest",
		})

		s.Run("returns OK", func() {
			s.NoError(err)
			s.False(toolResult.IsError)
		})

		s.Run("returns container ID", func() {
			text := toolResult.Content[0].(mcp.TextContent).Text
			s.Contains(text, "container-shortname", "should contain the container ID")
		})
	})

	s.Run("container_run(imageName=example.com/org/image:tag) runs container", func() {
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
	})

	s.Run("container_run(ports=[...]) includes port mappings", func() {
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
			req := s.PopLastCapturedRequest("POST", "/libpod/containers/create")
			s.Require().NotNil(req, "create request should be captured")
			s.Contains(req.Body, `"host_port":8080`, "should have port 8080 mapping")
			s.Contains(req.Body, `"host_port":8082`, "should have port 8082 mapping")
			s.Contains(req.Body, `"host_port":8443`, "should have port 8443 mapping")
		})
	})

	s.Run("container_run(environment=[...]) includes environment variables", func() {
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
			req := s.PopLastCapturedRequest("POST", "/libpod/containers/create")
			s.Require().NotNil(req, "create request should be captured")
			s.Contains(req.Body, `"KEY":"VALUE"`, "should have KEY=VALUE env var")
			s.Contains(req.Body, `"FOO":"BAR"`, "should have FOO=BAR env var")
		})
	})
}
