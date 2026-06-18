package toolsets_test

import (
	"testing"

	"github.com/manusa/podman-mcp-server/pkg/api"
	"github.com/manusa/podman-mcp-server/pkg/toolsets"
	"github.com/stretchr/testify/assert"
)

type fakeToolset struct{ name string }

func (f fakeToolset) GetName() string            { return f.name }
func (f fakeToolset) GetDescription() string     { return "fake" }
func (f fakeToolset) GetTools() []api.ServerTool { return nil }

func TestRegisterAndRetrieve(t *testing.T) {
	toolsets.Register(fakeToolset{name: "a-test"})
	toolsets.Register(fakeToolset{name: "b-test"})

	t.Run("ToolsetFromString finds registered", func(t *testing.T) {
		assert.NotNil(t, toolsets.ToolsetFromString("a-test"))
		assert.NotNil(t, toolsets.ToolsetFromString("b-test"))
	})

	t.Run("ToolsetFromString returns nil for unknown", func(t *testing.T) {
		assert.Nil(t, toolsets.ToolsetFromString("c-test"))
	})

	t.Run("Toolsets returns sorted by name", func(t *testing.T) {
		all := toolsets.Toolsets()
		assert.Equal(t, "a-test", all[0].GetName())
	})
}
