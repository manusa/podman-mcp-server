package mcp

import (
	"context"

	"github.com/manusa/podman-mcp-server/pkg/api"
)

func initContainerTools() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "container_inspect",
				Description: "Displays the low-level information and configuration of a Docker or Podman container with the specified container ID or name",
				Annotations: api.ToolAnnotations{
					Title:           "Container: Inspect",
					ReadOnlyHint:    ptr(true),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"name": {
							Type:        "string",
							Description: "Docker or Podman container ID or name to display the information",
						},
					},
					Required: []string{"name"},
				},
			},
			Handler: containerInspect,
		},
		{
			Tool: api.Tool{
				Name:        "container_list",
				Description: "Prints out information about the running Docker or Podman containers",
				Annotations: api.ToolAnnotations{
					Title:           "Container: List",
					ReadOnlyHint:    ptr(true),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
				},
			},
			Handler: containerList,
		},
		{
			Tool: api.Tool{
				Name:        "container_logs",
				Description: "Displays the logs of a Docker or Podman container with the specified container ID or name",
				Annotations: api.ToolAnnotations{
					Title:           "Container: Logs",
					ReadOnlyHint:    ptr(true),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"name": {
							Type:        "string",
							Description: "Docker or Podman container ID or name to display the logs",
						},
					},
					Required: []string{"name"},
				},
			},
			Handler: containerLogs,
		},
		{
			Tool: api.Tool{
				Name:        "container_remove",
				Description: "Removes a Docker or Podman container with the specified container ID or name (rm)",
				Annotations: api.ToolAnnotations{
					Title:           "Container: Remove",
					ReadOnlyHint:    ptr(false),
					DestructiveHint: ptr(true),
					IdempotentHint:  ptr(false),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"name": {
							Type:        "string",
							Description: "Docker or Podman container ID or name to remove",
						},
					},
					Required: []string{"name"},
				},
			},
			Handler: containerRemove,
		},
		{
			Tool: api.Tool{
				Name:        "container_run",
				Description: "Runs a Docker or Podman container with the specified image name",
				Annotations: api.ToolAnnotations{
					Title:           "Container: Run",
					ReadOnlyHint:    ptr(false),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(false),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"imageName": {
							Type:        "string",
							Description: "Docker or Podman container image name to run",
						},
						"ports": {
							Type:        "array",
							Description: "Port mappings to expose on the host. Format: <hostPort>:<containerPort>. Example: 8080:80. (Optional, add only to expose ports)",
							Items: &api.Property{
								Type: "string",
							},
						},
						"environment": {
							Type:        "array",
							Description: "Environment variables to set in the container. Format: <key>=<value>. Example: FOO=bar. (Optional, add only to set environment variables)",
							Items: &api.Property{
								Type: "string",
							},
						},
					},
					Required: []string{"imageName"},
				},
			},
			Handler: containerRun,
		},
		{
			Tool: api.Tool{
				Name:        "container_stop",
				Description: "Stops a Docker or Podman running container with the specified container ID or name",
				Annotations: api.ToolAnnotations{
					Title:           "Container: Stop",
					ReadOnlyHint:    ptr(false),
					DestructiveHint: ptr(false),
					IdempotentHint:  ptr(true),
					OpenWorldHint:   ptr(false),
				},
				InputSchema: api.InputSchema{
					Type: "object",
					Properties: map[string]api.Property{
						"name": {
							Type:        "string",
							Description: "Docker or Podman container ID or name to stop",
						},
					},
					Required: []string{"name"},
				},
			},
			Handler: containerStop,
		},
	}
}

func containerInspect(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	name, err := params.RequiredString("name")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	result, err := params.Podman.ContainerInspect(name)
	return api.NewToolCallResult(result, err), nil
}

func containerList(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	result, err := params.Podman.ContainerList()
	return api.NewToolCallResult(result, err), nil
}

func containerLogs(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	name, err := params.RequiredString("name")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	result, err := params.Podman.ContainerLogs(name)
	return api.NewToolCallResult(result, err), nil
}

func containerRemove(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	name, err := params.RequiredString("name")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	result, err := params.Podman.ContainerRemove(name)
	return api.NewToolCallResult(result, err), nil
}

func containerRun(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	imageName, err := params.RequiredString("imageName")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	portMappings := params.GetPortMappings("ports")
	envVariables := params.GetStringArray("environment")
	result, err := params.Podman.ContainerRun(imageName, portMappings, envVariables)
	return api.NewToolCallResult(result, err), nil
}

func containerStop(_ context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	name, err := params.RequiredString("name")
	if err != nil {
		return api.NewToolCallResult("", err), nil
	}
	result, err := params.Podman.ContainerStop(name)
	return api.NewToolCallResult(result, err), nil
}
