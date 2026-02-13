package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

// VolumeSuite tests volume tools using the mock Podman API server.
// These tests use the real podman CLI binary communicating with a mocked backend.
type VolumeSuite struct {
	test.McpSuite
}

func TestVolumeSuite(t *testing.T) {
	suite.Run(t, new(VolumeSuite))
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

		var volumes []test.VolumeResponse
		s.Require().NoError(json.Unmarshal([]byte(text), &volumes), "output should be valid JSON")

		s.Require().Len(volumes, 2)

		volumesByName := make(map[string]test.VolumeResponse)
		for _, val := range volumes {
			volumesByName[val.Name] = val
		}

		s.Require().Contains(volumesByName, "my-volume", "should contain my-volume volume")
		s.Require().Contains(volumesByName, "data-volume", "should contain data-volume volume")

		myVolume := volumesByName["my-volume"]
		s.Equal("local", myVolume.Driver)
		s.Contains(myVolume.Mountpoint, "/my-volume/_data")
		s.Equal(map[string]string{"app": "test"}, myVolume.Labels)

		dataVolume := volumesByName["data-volume"]
		s.Equal("local", dataVolume.Driver)
		s.Contains(dataVolume.Mountpoint, "/data-volume/_data")
		s.Nil(dataVolume.Labels, "data-volume should have no labels")
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
