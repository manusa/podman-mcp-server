package mcp

import (
	"context"

	"github.com/manusa/podman-mcp-server/pkg/api"
)

func initImageTools() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "image_build",
				Description: "Build a Docker or Podman image from a Dockerfile, Podmanfile, or Containerfile",
				Annotations: api.ToolAnnotations{
					Title:           "Image: Build",
					ReadOnlyHint:    ptr(false),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(false),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"containerFile": {
							Type:        "string",
							Description: "The absolute path to the Dockerfile, Podmanfile, or Containerfile to build the image from",
						},
						"imageName": {
							Type:        "string",
							Description: "Specifies the name which is assigned to the resulting image if the build process completes successfully (--tag, -t)",
						},
					},
					Required: []string{"containerFile"},
				},
			},
			Handler: imageBuild,
		},
		{
			Tool: api.Tool{
				Name:        "image_list",
				Description: "List the Docker or Podman images on the local machine",
				Annotations: api.ToolAnnotations{
					Title:           "Image: List",
					ReadOnlyHint:    ptr(true),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
				},
			},
			Handler: imageList,
		},
		{
			Tool: api.Tool{
				Name:        "image_pull",
				Description: "Copies (pulls) a Docker or Podman container image from a registry onto the local machine storage",
				Annotations: api.ToolAnnotations{
					Title:           "Image: Pull",
					ReadOnlyHint:    ptr(false),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(true),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"imageName": {
							Type:        "string",
							Description: "Docker or Podman container image name to pull",
						},
					},
					Required: []string{"imageName"},
				},
			},
			Handler: imagePull,
		},
		{
			Tool: api.Tool{
				Name:        "image_push",
				Description: "Pushes a Docker or Podman container image, manifest list or image index from local machine storage to a registry",
				Annotations: api.ToolAnnotations{
					Title:           "Image: Push",
					ReadOnlyHint:    ptr(false),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(true),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"imageName": {
							Type:        "string",
							Description: "Docker or Podman container image name to push",
						},
					},
					Required: []string{"imageName"},
				},
			},
			Handler: imagePush,
		},
		{
			Tool: api.Tool{
				Name:        "image_remove",
				Description: "Removes a Docker or Podman image from the local machine storage",
				Annotations: api.ToolAnnotations{
					Title:           "Image: Remove",
					ReadOnlyHint:    ptr(false),
					DestructiveHint: ptr(true),
					IdempotentHint:  ptr(false),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"imageName": {
							Type:        "string",
							Description: "Docker or Podman container image name to remove",
						},
					},
					Required: []string{"imageName"},
				},
			},
			Handler: imageRemove,
		},
	}
}

func imageBuild(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	containerFile, err := params.RequiredString("containerFile")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	imageName := params.GetString("imageName", "")
	result, err := params.Podman.ImageBuild(containerFile, imageName)
	return api.NewToolCallResult(result, err), nil
}

func imageList(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	result, err := params.Podman.ImageList()
	return api.NewToolCallResult(result, err), nil
}

func imagePull(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	imageName, err := params.RequiredString("imageName")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	result, err := params.Podman.ImagePull(imageName)
	return api.NewToolCallResult(result, err), nil
}

func imagePush(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	imageName, err := params.RequiredString("imageName")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	result, err := params.Podman.ImagePush(imageName)
	return api.NewToolCallResult(result, err), nil
}

func imageRemove(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	imageName, err := params.RequiredString("imageName")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	result, err := params.Podman.ImageRemove(imageName)
	return api.NewToolCallResult(result, err), nil
}
