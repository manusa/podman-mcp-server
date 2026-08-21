package network

import (
	"context"

	"github.com/manusa/podman-mcp-server/pkg/api"
	"github.com/manusa/podman-mcp-server/pkg/toolsets"
)

type toolset struct{}

func (t toolset) GetName() string        { return "network" }
func (t toolset) GetDescription() string { return "Network management tools" }

func (t toolset) GetTools() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "network_list",
				Description: "List all the available Docker or Podman networks",
				Annotations: api.ToolAnnotations{
					Title:           "Network: List",
					ReadOnlyHint:    api.Ptr(true),
					DestructiveHint: api.Ptr(false),
					IdempotentHint:  api.Ptr(true),
					OpenWorldHint:   api.Ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
				},
			},
			Handler: networkList,
		},
	}
}

func init() {
	toolsets.Register(toolset{})
}

func networkList(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	result, err := params.Podman.NetworkList()
	return api.NewToolCallResult(result, err), nil
}
