package mcp

import (
	"context"

	"github.com/manusa/podman-mcp-server/pkg/api"
)

func initVolumeTools() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "volume_list",
				Description: "List all the available Docker or Podman volumes",
				Annotations: api.ToolAnnotations{
					Title:           "Volume: List",
					ReadOnlyHint:    ptr(true),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
				},
			},
			Handler: volumeList,
		},
	}
}

func volumeList(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	result, err := params.Podman.VolumeList()
	return api.NewToolCallResult(result, err), nil
}
