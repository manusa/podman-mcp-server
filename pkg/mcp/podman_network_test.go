package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"strings"
	"testing"
)

func TestNetworkList(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		toolResult, err := c.callTool("network_list", map[string]interface{}{})
		t.Run("network_list returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
		})
		t.Run("network_list lists all available networks", func(t *testing.T) {
			if !strings.HasPrefix(toolResult.Content[0].(mcp.TextContent).Text, "podman network ls") {
				t.Fatalf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
	})
}
