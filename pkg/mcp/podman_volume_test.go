package mcp_test

import (
	"regexp"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
	"github.com/manusa/podman-mcp-server/pkg/config"
)

// VolumeSuite tests volume tools using the mock Podman API server.
// These tests use the real podman CLI binary communicating with a mocked backend.
type VolumeSuite struct {
	test.McpSuite
}

func TestVolumeSuiteWithAllImplementations(t *testing.T) {
	for _, impl := range test.AvailableImplementations() {
		t.Run(impl, func(t *testing.T) {
			suite.Run(t, &VolumeSuite{
				McpSuite: test.McpSuite{Config: config.Config{PodmanImpl: impl}},
			})
		})
	}
}

func (s *VolumeSuite) TestVolumeList() {
	s.WithVolumeList([]test.VolumeResponse{
		{
			Name:       "my-volume",
			Driver:     "local",
			Mountpoint: "/var/lib/containers/storage/volumes/my-volume/_data",
			CreatedAt:  "2024-01-01T00:00:00Z",
			Labels:     map[string]string{"app": "test"},
			Scope:      "local",
		},
		{
			Name:       "data-volume",
			Driver:     "local",
			Mountpoint: "/var/lib/containers/storage/volumes/data-volume/_data",
			CreatedAt:  "2024-01-02T00:00:00Z",
			Scope:      "local",
		},
	})

	toolResult, err := s.CallTool("volume_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns volume data with expected format", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text

		expectedHeaders := regexp.MustCompile(`(?m)^DRIVER\s+VOLUME NAME\s*$`)
		s.Regexpf(expectedHeaders, text, "expected headers not found in output:\n%s", text)

		s.Contains(text, "my-volume", "should contain volume name")
		s.Contains(text, "data-volume", "should contain second volume name")
		s.Contains(text, "local", "should contain driver")
	})

	s.Run("mock server received volume list request", func() {
		s.True(s.MockServer.HasRequest("GET", "/libpod/volumes/json"))
	})
}

func (s *VolumeSuite) TestVolumeListEmpty() {
	s.WithVolumeList([]test.VolumeResponse{})

	toolResult, err := s.CallTool("volume_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty or headers-only output", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		// Some podman versions print headers even when empty, others don't
		// Just verify no volume data is present
		s.NotContains(text, "my-volume", "should not contain volume data")
	})
}
