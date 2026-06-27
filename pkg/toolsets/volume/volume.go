package volume

import (
	"context"

	"github.com/manusa/podman-mcp-server/pkg/api"
	"github.com/manusa/podman-mcp-server/pkg/toolsets"
)

type toolset struct{}

func (t toolset) GetName() string        { return "volume" }
func (t toolset) GetDescription() string { return "Volume management tools" }

func (t toolset) GetTools() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "volume_list",
				Description: "List all the available Docker or Podman volumes",
				Annotations: api.ToolAnnotations{
					Title:           "Volume: List",
					ReadOnlyHint:    api.Ptr(true),
					DestructiveHint: api.Ptr(false),
					IdempotentHint:  api.Ptr(true),
					OpenWorldHint:   api.Ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
				},
			},
			Handler: volumeList,
		},
	}
}

func init() {
	toolsets.Register(toolset{})
}

func volumeList(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	result, err := params.Podman.VolumeList()
	return api.NewToolCallResult(result, err), nil
}
