package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"os"
	"path"
	"strings"
	"testing"
)

func TestImageBuild(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		toolResult, err := c.callTool("image_build", map[string]interface{}{
			"containerFileContent": "FROM scratch\nRUN echo hello",
		})
		t.Run("image_build returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if !strings.HasPrefix(toolResult.Content[0].(mcp.TextContent).Text, "podman build ") {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}

			if !strings.Contains(toolResult.Content[0].(mcp.TextContent).Text, path.Join(os.TempDir(), "Containerfile")) {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
		toolResult, err = c.callTool("image_build", map[string]interface{}{
			"containerFileContent": "FROM scratch\nRUN echo hello",
			"imageName":            "example.com/org/image:tag",
		})
		t.Run("image_build with imageName returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if !strings.HasPrefix(toolResult.Content[0].(mcp.TextContent).Text, "podman build -t example.com/org/image:tag") {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}

			if !strings.Contains(toolResult.Content[0].(mcp.TextContent).Text, path.Join(os.TempDir(), "Containerfile")) {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
	})
}

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
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if !strings.HasPrefix(toolResult.Content[0].(mcp.TextContent).Text, "podman image pull example.com/org/image:tag") {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
			if !strings.HasSuffix(toolResult.Content[0].(mcp.TextContent).Text, "example.com/org/image:tag pulled successfully") {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
	})
}

func TestImagePush(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		toolResult, err := c.callTool("image_push", map[string]interface{}{
			"imageName": "example.com/org/image:tag",
		})
		t.Run("image_push returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if !strings.HasPrefix(toolResult.Content[0].(mcp.TextContent).Text, "podman image push example.com/org/image:tag") {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
			if !strings.HasSuffix(toolResult.Content[0].(mcp.TextContent).Text, "example.com/org/image:tag pushed successfully") {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
	})
}

func TestImageRemove(t *testing.T) {
	testCase(t, func(c *mcpContext) {
		toolResult, err := c.callTool("image_remove", map[string]interface{}{
			"imageName": "example.com/org/image:tag",
		})
		t.Run("image_remove returns OK", func(t *testing.T) {
			if err != nil {
				t.Fatalf("call tool failed %v", err)
			}
			if toolResult.IsError {
				t.Fatalf("call tool failed")
			}
			if !strings.HasPrefix(toolResult.Content[0].(mcp.TextContent).Text, "podman image rm example.com/org/image:tag") {
				t.Errorf("unexpected result %v", toolResult.Content[0].(mcp.TextContent).Text)
			}
		})
	})
}
