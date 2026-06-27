package toolsets

import (
	"fmt"
	"slices"
	"strings"

	"github.com/manusa/podman-mcp-server/pkg/api"
)

var registry = make(map[string]api.Toolset)

// Register adds a toolset to the global registry. Panics on duplicate name,
// which is a programming error surfaced at init() time.
func Register(toolset api.Toolset) {
	name := toolset.GetName()
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("toolset already registered for name: %s", name))
	}
	registry[name] = toolset
}

// Toolsets returns all registered toolsets sorted by name for deterministic order.
func Toolsets() []api.Toolset {
	result := make([]api.Toolset, 0, len(registry))
	for _, ts := range registry {
		result = append(result, ts)
	}
	slices.SortFunc(result, func(a, b api.Toolset) int {
		return strings.Compare(a.GetName(), b.GetName())
	})
	return result
}

// ToolsetFromString returns the toolset with the given name, or nil if none.
func ToolsetFromString(name string) api.Toolset {
	return registry[strings.TrimSpace(name)]
}
