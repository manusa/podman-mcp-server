// Package api provides internal types for tool definitions decoupled from any specific MCP SDK.
package api

import (
	"context"

	"github.com/manusa/podman-mcp-server/pkg/podman"
)

// ServerTool represents a tool with its handler.
type ServerTool struct {
	Tool    Tool
	Handler ToolHandlerFunc
}

// ToolHandlerFunc is the function signature for tool handlers.
type ToolHandlerFunc func(ctx context.Context, params ToolHandlerParams) (*ToolCallResult, error)

// ToolHandlerParams contains all parameters passed to a tool handler.
type ToolHandlerParams struct {
	Podman    podman.Podman
	Arguments map[string]any
}

// Tool represents a tool definition.
type Tool struct {
	Name        string
	Description string
	Annotations ToolAnnotations
	InputSchema InputSchema
}

// ToolAnnotations contains MCP tool hints.
type ToolAnnotations struct {
	Title           string
	ReadOnlyHint    *bool
	DestructiveHint *bool
	IdempotentHint  *bool
	OpenWorldHint   *bool
}

// InputSchema defines the JSON schema for tool input parameters.
type InputSchema struct {
	Type       string
	Properties map[string]Property
	Required   []string
}

// Property defines a single property in the input schema.
type Property struct {
	Type        string
	Description string
	Items       *Property // for array types
}

// ToolCallResult represents the result of a tool call.
type ToolCallResult struct {
	Content string
	Error   error
}

// NewToolCallResult creates a new tool call result.
func NewToolCallResult(content string, err error) *ToolCallResult {
	return &ToolCallResult{
		Content: content,
		Error:   err,
	}
}
