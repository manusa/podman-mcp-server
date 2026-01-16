package mcp

import (
	"context"

	"github.com/manusa/podman-mcp-server/pkg/api"
)

func initNetworkTools() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "network_list",
				Description: "List all the available Docker or Podman networks",
				Annotations: api.ToolAnnotations{
					Title:           "Network: List",
					ReadOnlyHint:    ptr(true),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
				},
			},
			Handler: networkList,
		},
	}
}

func networkList(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	result, err := params.Podman.NetworkList()
	return api.NewToolCallResult(result, err), nil
}
