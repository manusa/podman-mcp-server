package mcp_test

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
	"github.com/manusa/podman-mcp-server/pkg/config"
)

// JSONOutputSuite tests list tools with JSON output format configured.
// These tests verify that when OutputFormat is set to "json", the tools
// return valid JSON instead of human-readable table format.
//
// Note on JSON key casing: The Podman API uses inconsistent key casing across
// different resource types. This reflects the actual API behavior:
//   - Containers/Images/Volumes: PascalCase keys (e.g., "Id", "Name", "Image")
//   - Networks: snake_case keys (e.g., "id", "name", "driver")
//
// These tests verify the actual API response format, not normalized keys.
type JSONOutputSuite struct {
	test.McpSuite
}

func TestJSONOutputSuiteWithAllImplementations(t *testing.T) {
	for _, impl := range test.AvailableImplementations() {
		t.Run(impl, func(t *testing.T) {
			suite.Run(t, &JSONOutputSuite{
				McpSuite: test.McpSuite{Config: config.Config{
					OutputFormat: "json",
					PodmanImpl:   impl,
				}},
			})
		})
	}
}

func (s *JSONOutputSuite) TestContainerListJSON() {
	s.WithContainerList([]test.ContainerListResponse{
		{
			ID:        "abc123def456",
			Names:     []string{"test-container-1"},
			Image:     "docker.io/library/nginx:latest",
			ImageID:   "sha256:abc123",
			State:     "running",
			Status:    "Up 2 hours",
			Created:   "2024-01-01T00:00:00Z",
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

	s.Run("returns valid JSON array", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text

		var containers []map[string]interface{}
		s.Require().NoError(json.Unmarshal([]byte(text), &containers), "output should be valid JSON array")
		s.Require().Len(containers, 2)

		// Containers use PascalCase keys (e.g., "Id", "Image", "State")
		containersByID := make(map[string]map[string]interface{})
		for _, c := range containers {
			containersByID[c["Id"].(string)] = c
		}

		s.Run("contains first container data", func() {
			c := containersByID["abc123def456"]
			s.Require().NotNil(c)
			s.Equal("docker.io/library/nginx:latest", c["Image"])
			s.Equal("running", c["State"])
			names := c["Names"].([]interface{})
			s.Contains(names, "test-container-1")
		})

		s.Run("contains second container data", func() {
			c := containersByID["xyz789ghi012"]
			s.Require().NotNil(c)
			s.Equal("docker.io/library/redis:alpine", c["Image"])
			s.Equal("exited", c["State"])
			names := c["Names"].([]interface{})
			s.Contains(names, "test-container-2")
		})
	})
}

func (s *JSONOutputSuite) TestContainerListEmptyJSON() {
	s.WithContainerList([]test.ContainerListResponse{})

	toolResult, err := s.CallTool("container_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty JSON array", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.JSONEq(`[]`, text)
	})
}

func (s *JSONOutputSuite) TestContainerListErrorJSON() {
	s.WithError("GET", "/libpod/containers/json", "/containers/json",
		500, "internal server error")

	toolResult, err := s.CallTool("container_list", map[string]interface{}{})

	s.Run("returns tool error", func() {
		s.NoError(err)
		s.True(toolResult.IsError)
	})

	s.Run("error message is not empty", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.NotEmpty(text)
	})
}

func (s *JSONOutputSuite) TestImageListJSON() {
	s.WithImageList([]test.ImageListResponse{
		{
			ID:          "sha256:abc123def456",
			RepoTags:    []string{"docker.io/library/nginx:latest"},
			RepoDigests: []string{"docker.io/library/nginx@sha256:abc123"},
			Size:        142000000,
			Created:     1704067200,
		},
		{
			ID:          "sha256:xyz789ghi012",
			RepoTags:    []string{"docker.io/library/redis:alpine"},
			RepoDigests: []string{"docker.io/library/redis@sha256:xyz789"},
			Size:        37000000,
			Created:     1704067200,
		},
	})

	toolResult, err := s.CallTool("image_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns valid JSON array", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text

		var images []map[string]interface{}
		s.Require().NoError(json.Unmarshal([]byte(text), &images), "output should be valid JSON array")
		s.Require().Len(images, 2)

		// Images use PascalCase keys (e.g., "Id", "RepoTags", "Size")
		imagesByID := make(map[string]map[string]interface{})
		for _, img := range images {
			imagesByID[img["Id"].(string)] = img
		}

		s.Run("contains nginx image", func() {
			img := imagesByID["sha256:abc123def456"]
			s.Require().NotNil(img)
			repoDigests := img["RepoDigests"].([]interface{})
			s.Contains(repoDigests[0], "nginx")
		})

		s.Run("contains redis image", func() {
			img := imagesByID["sha256:xyz789ghi012"]
			s.Require().NotNil(img)
			repoDigests := img["RepoDigests"].([]interface{})
			s.Contains(repoDigests[0], "redis")
		})
	})
}

func (s *JSONOutputSuite) TestImageListEmptyJSON() {
	s.WithImageList([]test.ImageListResponse{})

	toolResult, err := s.CallTool("image_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty JSON array", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.JSONEq(`[]`, text)
	})
}

func (s *JSONOutputSuite) TestImageListErrorJSON() {
	s.WithError("GET", "/libpod/images/json", "/images/json",
		500, "internal server error")

	toolResult, err := s.CallTool("image_list", map[string]interface{}{})

	s.Run("returns tool error", func() {
		s.NoError(err)
		s.True(toolResult.IsError)
	})

	s.Run("error message is not empty", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.NotEmpty(text)
	})
}

func (s *JSONOutputSuite) TestNetworkListJSON() {
	s.WithNetworkList([]test.NetworkListResponse{
		{
			Name:             "podman",
			ID:               "abc123def456",
			Driver:           "bridge",
			NetworkInterface: "podman0",
			DNSEnabled:       true,
			Subnets: []test.Subnet{
				{Subnet: "10.88.0.0/16", Gateway: "10.88.0.1"},
			},
		},
		{
			Name:             "my-network",
			ID:               "xyz789ghi012",
			Driver:           "bridge",
			NetworkInterface: "",
			DNSEnabled:       true,
			Subnets: []test.Subnet{
				{Subnet: "10.89.0.0/24", Gateway: "10.89.0.1"},
			},
		},
	})

	toolResult, err := s.CallTool("network_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns valid JSON array", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text

		var networks []map[string]interface{}
		s.Require().NoError(json.Unmarshal([]byte(text), &networks), "output should be valid JSON array")
		s.Require().Len(networks, 2)

		// Networks use snake_case keys (e.g., "id", "name", "driver")
		networksByName := make(map[string]map[string]interface{})
		for _, net := range networks {
			networksByName[net["name"].(string)] = net
		}

		s.Run("contains podman network", func() {
			net := networksByName["podman"]
			s.Require().NotNil(net)
			s.Equal("bridge", net["driver"])
			s.Equal("abc123def456", net["id"])
		})

		s.Run("contains custom network", func() {
			net := networksByName["my-network"]
			s.Require().NotNil(net)
			s.Equal("bridge", net["driver"])
			s.Equal("xyz789ghi012", net["id"])
		})
	})
}

func (s *JSONOutputSuite) TestNetworkListEmptyJSON() {
	s.WithNetworkList([]test.NetworkListResponse{})

	toolResult, err := s.CallTool("network_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty JSON array", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.JSONEq(`[]`, text)
	})
}

func (s *JSONOutputSuite) TestNetworkListErrorJSON() {
	s.WithError("GET", "/libpod/networks/json", "/networks",
		500, "internal server error")

	toolResult, err := s.CallTool("network_list", map[string]interface{}{})

	s.Run("returns tool error", func() {
		s.NoError(err)
		s.True(toolResult.IsError)
	})

	s.Run("error message is not empty", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.NotEmpty(text)
	})
}

func (s *JSONOutputSuite) TestVolumeListJSON() {
	s.WithVolumeList([]test.VolumeResponse{
		{
			Name:       "my-volume",
			Driver:     "local",
			Mountpoint: "/var/lib/containers/storage/volumes/my-volume/_data",
			Labels:     map[string]string{"app": "test"},
		},
		{
			Name:       "data-volume",
			Driver:     "local",
			Mountpoint: "/var/lib/containers/storage/volumes/data-volume/_data",
			Labels:     nil,
		},
	})

	toolResult, err := s.CallTool("volume_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns valid JSON array", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text

		var volumes []map[string]interface{}
		s.Require().NoError(json.Unmarshal([]byte(text), &volumes), "output should be valid JSON array")
		s.Require().Len(volumes, 2)

		// Volumes use PascalCase keys (e.g., "Name", "Driver", "Mountpoint")
		volumesByName := make(map[string]map[string]interface{})
		for _, vol := range volumes {
			volumesByName[vol["Name"].(string)] = vol
		}

		s.Run("contains my-volume", func() {
			vol := volumesByName["my-volume"]
			s.Require().NotNil(vol)
			s.Equal("local", vol["Driver"])
			s.Contains(vol["Mountpoint"], "my-volume")
		})

		s.Run("contains data-volume", func() {
			vol := volumesByName["data-volume"]
			s.Require().NotNil(vol)
			s.Equal("local", vol["Driver"])
			s.Contains(vol["Mountpoint"], "data-volume")
		})
	})
}

func (s *JSONOutputSuite) TestVolumeListEmptyJSON() {
	s.WithVolumeList([]test.VolumeResponse{})

	toolResult, err := s.CallTool("volume_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty JSON array", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.JSONEq(`[]`, text)
	})
}

func (s *JSONOutputSuite) TestVolumeListErrorJSON() {
	s.WithError("GET", "/libpod/volumes/json", "/volumes",
		500, "internal server error")

	toolResult, err := s.CallTool("volume_list", map[string]interface{}{})

	s.Run("returns tool error", func() {
		s.NoError(err)
		s.True(toolResult.IsError)
	})

	s.Run("error message is not empty", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.NotEmpty(text)
	})
}
