package mcp_test

import (
	"regexp"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
)

// NetworkSuite tests network tools using the mock Podman API server.
// These tests use the real podman CLI binary communicating with a mocked backend.
type NetworkSuite struct {
	test.McpSuite
}

func TestNetworkSuite(t *testing.T) {
	suite.Run(t, new(NetworkSuite))
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
		text := toolResult.Content[0].(mcp.TextContent).Text

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

func (s *NetworkSuite) TestNetworkListEmpty() {
	s.WithNetworkList([]test.NetworkListResponse{})

	toolResult, err := s.CallTool("network_list", map[string]interface{}{})

	s.Run("returns OK", func() {
		s.NoError(err)
		s.False(toolResult.IsError)
	})

	s.Run("returns empty or headers-only output", func() {
		text := toolResult.Content[0].(mcp.TextContent).Text
		// Some podman versions print headers even when empty, others don't
		// Just verify no network data is present
		s.NotContains(text, "my-network", "should not contain network data")
	})
}
