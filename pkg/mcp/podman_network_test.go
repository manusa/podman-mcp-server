package mcp_test

import (
	"encoding/json"
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

		var networks []test.NetworkListResponse
		s.Require().NoError(json.Unmarshal([]byte(text), &networks))

		s.Require().Len(networks, 2)

		networksByName := make(map[string]test.NetworkListResponse)
		for _, net := range networks {
			networksByName[net.Name] = net
		}

		s.Require().Contains(networksByName, "podman", "should contain podman network")
		s.Require().Contains(networksByName, "my-network", "should contain my-network network")

		podmanNetwork := networksByName["podman"]
		s.Equal("bridge", podmanNetwork.Driver)
		s.Equal("podman0", podmanNetwork.NetworkInterface)
		s.Equal("abc123def456", podmanNetwork.ID)
		s.True(podmanNetwork.DNSEnabled, "podman network should have DNS enabled")
		s.Require().NotEmpty(podmanNetwork.Subnets, "podman network should have subnets")
		s.Equal("10.88.0.0/16", podmanNetwork.Subnets[0].Subnet)

		myNetwork := networksByName["my-network"]
		s.Equal("bridge", myNetwork.Driver)
		s.Equal("", myNetwork.NetworkInterface)
		s.Equal("xyz789ghi012", myNetwork.ID)
		s.True(myNetwork.DNSEnabled, "my-network should have DNS enabled")
		s.Require().NotEmpty(myNetwork.Subnets, "my-network should have subnets")
		s.Equal("10.89.0.0/24", myNetwork.Subnets[0].Subnet)
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
