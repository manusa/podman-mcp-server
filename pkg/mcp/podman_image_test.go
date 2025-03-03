package mcp

import "testing"

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
			if toolResult.Content[0].(map[string]interface{})["text"].(string) != "" {
				t.Fatalf("unexpected content")
				return
			}
		})
	})
}
