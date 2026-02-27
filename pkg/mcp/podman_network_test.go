package mcp_test

import (
	"regexp"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
	"github.com/manusa/podman-mcp-server/pkg/config"
)

// NetworkSuite tests network tools using the mock Podman API server.
// These tests use the real podman CLI binary communicating with a mocked backend.
type NetworkSuite struct {
	test.McpSuite
}

func TestNetworkSuiteWithAllImplementations(t *testing.T) {
	for _, impl := range test.AvailableImplementations() {
		t.Run(impl, func(t *testing.T) {
			suite.Run(t, &NetworkSuite{
				McpSuite: test.McpSuite{Config: config.Config{PodmanImpl: impl}},
			})
		})
	}
}

func (s *NetworkSuite) TestNetworkList() {
	s.WithNetworkList([]test.NetworkListResponse{
		{
			Name:   "podman",
			ID:     "abc123def456",
			Driver: "bridge",
			Scope:  "local",
			Subnets: []test.Subnet{
				{Subnet: "10.88.0.0/16", Gateway: "10.88.0.1"},
			},
			DNSEnabled:       true,
			Internal:         false,
			NetworkInterface: "podman0",
		},
		{
			Name:   "my-network",
			ID:     "xyz789ghi012",
			Driver: "bridge",
			Scope:  "local",
			Subnets: []test.Subnet{
				{Subnet: "10.89.0.0/24", Gateway: "10.89.0.1"},
			},
			DNSEnabled: true,
			Internal:   false,
		},
	})

	toolResult, err := s.CallTool("network_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns network data with expected format", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text

		expectedHeaders := regexp.MustCompile(`(?m)^NETWORK ID\s+NAME\s+DRIVER\s*$`)
		s.Regexpf(expectedHeaders, text, "expected headers not found in output:\n%s", text)

		s.Contains(text, "podman", "should contain podman network")
		s.Contains(text, "my-network", "should contain custom network")
		s.Contains(text, "bridge", "should contain driver type")
	})

	s.Run("mock server received network list request", func() {
		s.True(s.MockServer.HasRequest("GET", "/libpod/networks/json"))
	})
}

func (s *NetworkSuite) TestNetworkListWithLongIds() {
	s.WithNetworkList([]test.NetworkListResponse{
		{
			Name:   "podman",
			ID:     "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			Driver: "bridge",
			Scope:  "local",
			Subnets: []test.Subnet{
				{Subnet: "10.88.0.0/16", Gateway: "10.88.0.1"},
			},
			DNSEnabled: true,
		},
		{
			Name:   "custom-net",
			ID:     "f1e2d3c4b5a6f7e8d9c0b1a2f3e4d5c6b7a8f9e0d1c2b3a4f5e6d7c8b9a0f1e2",
			Driver: "bridge",
			Scope:  "local",
			Subnets: []test.Subnet{
				{Subnet: "10.89.0.0/24", Gateway: "10.89.0.1"},
			},
			DNSEnabled: true,
		},
	})

	toolResult, err := s.CallTool("network_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("truncates long network IDs", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.Contains(text, "a1b2c3d4e5f6", "should contain truncated network ID")
		s.NotContains(text, "a1b2c3d4e5f6a7b8", "should not contain full ID beyond 12 chars")
	})

	s.Run("returns network data", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		s.Contains(text, "podman", "should contain podman network")
		s.Contains(text, "custom-net", "should contain custom network")
	})
}

func (s *NetworkSuite) TestNetworkListEmpty() {
	s.WithNetworkList([]test.NetworkListResponse{})

	toolResult, err := s.CallTool("network_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty or headers-only output", func() {
		text := toolResult.Content[0].(*mcp.TextContent).Text
		// Some podman versions print headers even when empty, others don't
		// Just verify no network data is present
		s.NotContains(text, "my-network", "should not contain network data")
	})
}
