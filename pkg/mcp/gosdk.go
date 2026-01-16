// Package mcp provides the MCP server implementation using the official Go SDK.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manusa/podman-mcp-server/pkg/api"
	"github.com/manusa/podman-mcp-server/pkg/podman"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ServerToolToGoSdkTool converts an internal ServerTool to go-sdk format.
func ServerToolToGoSdkTool(p podman.Podman, tool api.ServerTool) (*mcp.Tool, mcp.ToolHandler, error) {
	goSdkTool := &mcp.Tool{
		Name:        tool.Tool.Name,
		Description: tool.Tool.Description,
		Annotations: apiAnnotationsToSdk(tool.Tool.Annotations),
		InputSchema: apiInputSchemaToSdk(tool.Tool.InputSchema),
	}

	goSdkHandler := func(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, err := parseToolCallArguments(request)
		if err != nil {
			return nil, fmt.Errorf("%v for tool %s", err, tool.Tool.Name)
		}

		params := api.ToolHandlerParams{
			Podman:    p,
			Arguments: arguments,
		}

		result, err := tool.Handler(ctx, params)
		if err != nil {
			return nil, err
		}
		return newGoSdkTextResult(result.Content, result.Error), nil
	}

	return goSdkTool, goSdkHandler, nil
}

// apiAnnotationsToSdk converts internal ToolAnnotations to go-sdk format.
func apiAnnotationsToSdk(ann api.ToolAnnotations) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		Title:           ann.Title,
		ReadOnlyHint:    derefBool(ann.ReadOnlyHint, false),
		DestructiveHint: ann.DestructiveHint,
		IdempotentHint:  derefBool(ann.IdempotentHint, false),
		OpenWorldHint:   ann.OpenWorldHint,
	}
}

// apiInputSchemaToSdk converts internal InputSchema to a map for go-sdk.
func apiInputSchemaToSdk(schema api.InputSchema) map[string]any {
	result := map[string]any{
		"type": schema.Type,
	}

	if len(schema.Properties) > 0 {
		properties := make(map[string]any)
		for name, prop := range schema.Properties {
			properties[name] = propertyToMap(prop)
		}
		result["properties"] = properties
	}

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	return result
}

// propertyToMap converts a Property to a map representation.
func propertyToMap(prop api.Property) map[string]any {
	result := map[string]any{
		"type": prop.Type,
	}
	if prop.Description != "" {
		result["description"] = prop.Description
	}
	if prop.Items != nil {
		result["items"] = propertyToMap(*prop.Items)
	}
	return result
}

// parseToolCallArguments extracts arguments from the go-sdk CallToolRequest.
func parseToolCallArguments(request *mcp.CallToolRequest) (map[string]any, error) {
	if request.Params == nil {
		return make(map[string]any), nil
	}
	if request.Params.Arguments == nil {
		return make(map[string]any), nil
	}
	var arguments map[string]any
	if err := json.Unmarshal(request.Params.Arguments, &arguments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
	}
	return arguments, nil
}

// newGoSdkTextResult creates an SDK-compatible CallToolResult.
func newGoSdkTextResult(content string, err error) *mcp.CallToolResult {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: err.Error(),
				},
			},
		}
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: content,
			},
		},
	}
}

// derefBool dereferences a bool pointer, returning the default value if nil.
func derefBool(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}
