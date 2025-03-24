package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"strings"
	"testing"
)

func TestImageList(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		c.withPodmanOutput(
			"REPOSITORY\tTAG\tDIGEST\tIMAGE ID\tCREATED\tSIZE",
			"docker.io/marcnuri/chuck-norris\nlatest\nsha256:1337\nb8f22a2b8410\n1 day ago\n37 MB",
		)
		toolResult, err := c.callTool("image_list", map[string]interface{}{})
		t.Run("image_list returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
				return
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
				return
			}
			if !strings.HasPrefix(toolResult.Content[0].(mcp.TextContent).Text, "podman images --digests") {
				t.Fatalf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
				return
			}
		})
	})
}

func TestImagePull(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		toolResult, err := c.callTool("image_pull", map[string]interface{}{
			"imageName": "example.com/org/image:tag",
		})
		t.Run("image_pull returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
				return
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
				return
			}
			if !strings.HasSuffix(toolResult.Content[0].(mcp.TextContent).Text, "example.com/org/image:tag pulled successfully") {
				t.Fatalf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
				return
			}
		})
	})
}
