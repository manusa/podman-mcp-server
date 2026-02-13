package podman_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/pkg/podman"
)

type RegistrySuite struct {
	suite.Suite
}

func TestRegistry(t *testing.T) {
	suite.Run(t, new(RegistrySuite))
}

func (s *RegistrySuite) TestImplementations() {
	s.Run("returns registered implementations", func() {
		impls := podman.Implementations()
		s.NotEmpty(impls, "should have at least one implementation")
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
		s.True(found, "should include CLI implementation")
	})
}

func (s *RegistrySuite) TestImplementationNames() {
	s.Run("returns sorted names", func() {
		names := podman.ImplementationNames()
		s.NotEmpty(names, "should have at least one implementation name")
		s.Contains(names, "cli", "should include 'cli'")
	})

	s.Run("names are sorted alphabetically", func() {
		names := podman.ImplementationNames()
		for i := 1; i < len(names); i++ {
			s.LessOrEqual(names[i-1], names[i], "names should be sorted")
		}
	})
}

func (s *RegistrySuite) TestImplementationFromString() {
	s.Run("finds CLI implementation by name", func() {
		impl := podman.ImplementationFromString("cli")
		s.NotNil(impl, "should find CLI implementation")
		s.Equal("cli", impl.Name())
	})

	s.Run("returns nil for unknown name", func() {
		impl := podman.ImplementationFromString("nonexistent")
		s.Nil(impl, "should return nil for unknown implementation")
	})
}

func (s *RegistrySuite) TestCLIImplementationMetadata() {
	impl := podman.ImplementationFromString("cli")
	s.Require().NotNil(impl, "CLI implementation must exist")

	s.Run("has correct name", func() {
		s.Equal("cli", impl.Name())
	})

	s.Run("has description", func() {
		s.NotEmpty(impl.Description(), "should have a description")
	})

	s.Run("has priority of 50", func() {
		s.Equal(50, impl.Priority())
	})

	s.Run("is available when podman is installed", func() {
		// This test assumes podman is available in the test environment
		// If it's not, this would be a different test
		s.True(impl.Available(), "should be available when podman is installed")
	})
}
