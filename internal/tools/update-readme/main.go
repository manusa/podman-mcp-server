package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/manusa/podman-mcp-server/pkg/mcp"
)

func main() {
	// Snyk reports false positive unless we flow the args through filepath.Clean and filepath.Localize in this specific order
	var err error
	localReadmePath := filepath.Clean(os.Args[1])
	localReadmePath, err = filepath.Localize(localReadmePath)
	if err != nil {
		panic(err)
	}
	readme, err := os.ReadFile(localReadmePath)
	if err != nil {
		panic(err)
	}

	// Get all tools
	tools := mcp.AllTools()

	// Group tools by category (extracted from tool name prefix)
	categories := make(map[string][]struct {
		name        string
		description string
		properties  map[string]struct {
			propType    string
			description string
		}
		required []string
	})

	for _, tool := range tools {
		// Extract category from tool name (e.g., "container_list" -> "container")
		parts := strings.SplitN(tool.Tool.Name, "_", 2)
		category := parts[0]

		props := make(map[string]struct {
			propType    string
			description string
		})
		for name, prop := range tool.Tool.InputSchema.Properties {
			props[name] = struct {
				propType    string
				description string
			}{
				propType:    prop.Type,
				description: prop.Description,
			}
		}

		categories[category] = append(categories[category], struct {
			name        string
			description string
			properties  map[string]struct {
				propType    string
				description string
			}
			required []string
		}{
			name:        tool.Tool.Name,
			description: tool.Tool.Description,
			properties:  props,
			required:    tool.Tool.InputSchema.Required,
		})
	}

	// Build the tools documentation
	toolsDocs := strings.Builder{}

	// Sort categories for consistent output
	for _, category := range slices.Sorted(maps.Keys(categories)) {
		categoryTools := categories[category]
		// Capitalize category name for display
		displayCategory := strings.ToUpper(category[:1]) + category[1:]
		toolsDocs.WriteString("<details>\n\n<summary>" + displayCategory + "</summary>\n\n")

		for _, tool := range categoryTools {
			toolsDocs.WriteString(fmt.Sprintf("- **%s** - %s\n", tool.name, tool.description))
			for _, propName := range slices.Sorted(maps.Keys(tool.properties)) {
				property := tool.properties[propName]
				toolsDocs.WriteString(fmt.Sprintf("  - `%s` (`%s`)", propName, property.propType))
				if slices.Contains(tool.required, propName) {
					toolsDocs.WriteString(" **(required)**")
				}
				toolsDocs.WriteString(fmt.Sprintf(" - %s\n", property.description))
			}
			toolsDocs.WriteString("\n")
		}
		toolsDocs.WriteString("</details>\n\n")
	}

	updated := replaceBetweenMarkers(
		string(readme),
		"<!-- AVAILABLE-TOOLS-START -->",
		"<!-- AVAILABLE-TOOLS-END -->",
		toolsDocs.String(),
	)

	if err := os.WriteFile(localReadmePath, []byte(updated), 0o644); err != nil {
		panic(err)
	}
}

func replaceBetweenMarkers(content, startMarker, endMarker, replacement string) string {
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content
	}
	endIdx := strings.Index(content, endMarker)
	if endIdx == -1 || endIdx <= startIdx {
		return content
	}
	return content[:startIdx+len(startMarker)] + "\n\n" + replacement + "\n" + content[endIdx:]
}
