package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"testing"
)

func TestTools(t *testing.T) {
	expectedNames := []string{
		"container_inspect",
		"container_list",
		"container_logs",
		"container_remove",
		"container_run",
		"container_stop",
		"image_list",
		"image_pull",
		"image_push",
		"image_remove",
		"network_list",
		"volume_list",
	}
	testCase(t, func(c *mcpContext) {
		tools, err := c.mcpClient.ListTools(c.ctx, mcp.ListToolsRequest{})
		t.Run("ListTools returns tools", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call ListTools failed %v", err)
			}
		})
		nameSet := make(map[string]bool)
		for _, tool := range tools.Tools {
			nameSet[tool.Name] = true
		}
		for _, name := range expectedNames {
			t.Run("ListTools has "+name+" tool", func(t *testing.T) {
				if nameSet[name] != true {
					t.Errorf("tool %s not found", name)
				}
			})
		}
	})
}
