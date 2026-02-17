package podman_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/pkg/config"
	"github.com/manusa/podman-mcp-server/pkg/podman"
)

// registryMockImplementation is a simple mock for registry tests.
type registryMockImplementation struct {
	name     string
	priority int
}

func (m *registryMockImplementation) Name() string        { return m.name }
func (m *registryMockImplementation) Description() string { return "Mock: " + m.name }
func (m *registryMockImplementation) Available() bool     { return true }
func (m *registryMockImplementation) Priority() int       { return m.priority }
func (m *registryMockImplementation) Initialize(_ config.Config) (podman.Podman, error) {
	return nil, nil
}

type RegistrySuite struct {
	suite.Suite
	savedImplementations []podman.Implementation
}

func TestRegistry(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}

func (s *RegistrySuite) SetupTest() {
	// Save current registry state before each test
	s.savedImplementations = podman.Implementations()
}

func (s *RegistrySuite) TearDownTest() {
	// Restore registry state after each test
	podman.Clear()
	for _, impl := range s.savedImplementations {
		podman.Register(impl)
	}
}

func (s *RegistrySuite) TestImplementations() {
	s.Run("returns registered implementations", func() {
		impls := podman.Implementations()
		s.NotEmpty(impls)
	})

	s.Run("includes CLI implementation", func() {
		impls := podman.Implementations()
		found := false
		for _, impl := range impls {
			if impl.Name() == "cli" {
				found = true
				break
			}
		}
		s.True(found)
	})

	s.Run("returns a copy not the original slice", func() {
		impls1 := podman.Implementations()
		impls2 := podman.Implementations()
		s.NotSame(&impls1[0], &impls2[0])
	})
}

func (s *RegistrySuite) TestImplementationNames() {
	s.Run("returns names including cli", func() {
		names := podman.ImplementationNames()
		s.NotEmpty(names)
		s.Contains(names, "cli")
	})

	s.Run("names are sorted alphabetically", func() {
		names := podman.ImplementationNames()
		for i := 1; i < len(names); i++ {
			s.LessOrEqual(names[i-1], names[i])
		}
	})
}

func (s *RegistrySuite) TestImplementationFromString() {
	s.Run("finds CLI implementation by name", func() {
		impl := podman.ImplementationFromString("cli")
		s.NotNil(impl)
		s.Equal("cli", impl.Name())
	})

	s.Run("returns nil for unknown name", func() {
		impl := podman.ImplementationFromString("nonexistent")
		s.Nil(impl)
	})

	s.Run("returns nil for empty name", func() {
		impl := podman.ImplementationFromString("")
		s.Nil(impl)
	})
}

func (s *RegistrySuite) TestRegister() {
	s.Run("adds implementation to registry", func() {
		initialCount := len(podman.Implementations())
		podman.Register(&registryMockImplementation{name: "test-register", priority: 1})
		s.Equal(initialCount+1, len(podman.Implementations()))
	})

	s.Run("registered implementation is findable", func() {
		podman.Register(&registryMockImplementation{name: "test-findable", priority: 1})
		impl := podman.ImplementationFromString("test-findable")
		s.NotNil(impl)
		s.Equal("test-findable", impl.Name())
	})
}

func (s *RegistrySuite) TestClear() {
	s.Run("removes all implementations", func() {
		// Verify we have implementations
		s.NotEmpty(podman.Implementations())

		// Clear
		podman.Clear()
		s.Empty(podman.Implementations())
	})

	s.Run("ImplementationNames returns empty after clear", func() {
		podman.Clear()
		s.Empty(podman.ImplementationNames())
	})

	s.Run("ImplementationFromString returns nil after clear", func() {
		podman.Clear()
		s.Nil(podman.ImplementationFromString("cli"))
	})
}

func (s *RegistrySuite) TestDefaultImplementation() {
	s.Run("returns cli when only cli is registered", func() {
		def := podman.DefaultImplementation()
		s.Equal("cli", def)
	})

	s.Run("returns empty string when registry is empty", func() {
		podman.Clear()
		def := podman.DefaultImplementation()
		s.Empty(def)
	})

	s.Run("returns highest priority implementation", func() {
		podman.Clear()
		podman.Register(&registryMockImplementation{name: "low", priority: 10})
		podman.Register(&registryMockImplementation{name: "high", priority: 100})
		podman.Register(&registryMockImplementation{name: "medium", priority: 50})

		def := podman.DefaultImplementation()
		s.Equal("high", def)
	})

	s.Run("returns first when priorities are equal", func() {
		podman.Clear()
		podman.Register(&registryMockImplementation{name: "first", priority: 50})
		podman.Register(&registryMockImplementation{name: "second", priority: 50})

		def := podman.DefaultImplementation()
		// When priorities are equal, the first registered wins
		s.Equal("first", def)
	})
}
