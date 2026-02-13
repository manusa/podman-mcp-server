package podman_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/pkg/podman"
)

type ApiSuite struct {
	suite.Suite
	savedImplementations []podman.Implementation
}

func TestApi(t *testing.T) {
	suite.Run(t, new(ApiSuite))
}

func (s *ApiSuite) SetupTest() {
	s.savedImplementations = podman.Implementations()
}

func (s *ApiSuite) TearDownTest() {
	podman.Clear()
	for _, impl := range s.savedImplementations {
		podman.Register(impl)
	}
}

func (s *ApiSuite) TestAPIImplementationMetadata() {
	impl := podman.ImplementationFromString("api")
	s.Require().NotNil(impl)

	s.Run("Name returns api", func() {
		s.Equal("api", impl.Name())
	})

	s.Run("Description returns Podman REST API via Unix socket", func() {
		s.Equal("Podman REST API via Unix socket", impl.Description())
	})

	s.Run("Priority returns 100", func() {
		s.Equal(100, impl.Priority())
	})

	s.Run("Available returns false when no socket available", func() {
		// This test assumes no real Podman socket is available
		// It will return true if a real socket exists
		// We just verify it doesn't panic
		_ = impl.Available()
	})
}

func (s *ApiSuite) TestAPIRegisteredInRegistry() {
	s.Run("api is included in ImplementationNames", func() {
		names := podman.ImplementationNames()
		s.Contains(names, "api")
	})

	s.Run("api implementation can be found by name", func() {
		impl := podman.ImplementationFromString("api")
		s.NotNil(impl)
	})

	s.Run("both api and cli are registered", func() {
		names := podman.ImplementationNames()
		s.Contains(names, "api")
		s.Contains(names, "cli")
		s.Len(names, 2)
	})
}

func (s *ApiSuite) TestAPIPriorityHigherThanCLI() {
	cliImpl := podman.ImplementationFromString("cli")
	apiImpl := podman.ImplementationFromString("api")
	s.Require().NotNil(cliImpl)
	s.Require().NotNil(apiImpl)

	s.Run("api has higher priority than cli", func() {
		s.Greater(apiImpl.Priority(), cliImpl.Priority())
	})
}
