package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"testing"
)

func TestContainerImagePull(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		toolResult, err := c.callTool("container_image_pull", map[string]interface{}{
			"imageName": "example.com/org/image:tag",
		})
		t.Run("container_image_pull returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
				return
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
				return
			}
			if toolResult.Content[0].(mcp.TextContent).Text != "example.com/org/image:tag pulled successfully" {
				t.Fatalf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
				return
			}
		})
	})
}
