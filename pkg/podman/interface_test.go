package podman_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/pkg/config"
	"github.com/manusa/podman-mcp-server/pkg/podman"
)

// mockImplementation is a test implementation for registry testing.
type mockImplementation struct {
	name        string
	description string
	available   bool
	priority    int
	initFunc    func(config.Config) (podman.Podman, error)
}

func (m *mockImplementation) Name() string        { return m.name }
func (m *mockImplementation) Description() string { return m.description }
func (m *mockImplementation) Available() bool     { return m.available }
func (m *mockImplementation) Priority() int       { return m.priority }
func (m *mockImplementation) Initialize(cfg config.Config) (podman.Podman, error) {
	if m.initFunc != nil {
		return m.initFunc(cfg)
	}
	return nil, nil
}

// InterfaceSuite tests the Podman interface and NewPodman function.
type InterfaceSuite struct {
	suite.Suite
	savedImplementations []podman.Implementation
}

func TestInterface(t *testing.T) {
	suite.Run(t, new(InterfaceSuite))
}

func (s *InterfaceSuite) SetupTest() {
	// Save current registry state
	s.savedImplementations = podman.Implementations()
}

func (s *InterfaceSuite) TearDownTest() {
	// Restore registry state
	podman.Clear()
	for _, impl := range s.savedImplementations {
		podman.Register(impl)
	}
}

func (s *InterfaceSuite) TestNewPodmanWithUnknownImplementation() {
	cfg := config.Config{PodmanImpl: "nonexistent-impl"}
	_, err := podman.NewPodman(cfg)

	s.Run("returns ErrUnknownImplementation", func() {
		s.Error(err)
		var unknownErr *podman.ErrUnknownImplementation
		s.ErrorAs(err, &unknownErr)
	})

	s.Run("error contains implementation name", func() {
		s.Contains(err.Error(), "nonexistent-impl")
	})

	s.Run("error lists valid options", func() {
		s.Contains(err.Error(), "valid options:")
	})
}

func (s *InterfaceSuite) TestNewPodmanWithUnavailableImplementation() {
	// Clear and register an unavailable mock
	podman.Clear()
	podman.Register(&mockImplementation{
		name:      "unavailable-mock",
		available: false,
		priority:  100,
	})

	cfg := config.Config{PodmanImpl: "unavailable-mock"}
	_, err := podman.NewPodman(cfg)

	s.Run("returns ErrImplementationNotAvailable", func() {
		s.Error(err)
		var notAvailErr *podman.ErrImplementationNotAvailable
		s.ErrorAs(err, &notAvailErr)
		s.Equal("unavailable-mock", notAvailErr.Name)
	})
}

func (s *InterfaceSuite) TestNewPodmanAutoDetectionWithNoAvailableImplementations() {
	// Clear and register only unavailable mocks
	podman.Clear()
	podman.Register(&mockImplementation{
		name:      "mock1",
		available: false,
		priority:  100,
	})
	podman.Register(&mockImplementation{
		name:      "mock2",
		available: false,
		priority:  50,
	})

	cfg := config.Config{} // Auto-detect (empty PodmanImpl)
	_, err := podman.NewPodman(cfg)

	s.Run("returns ErrNoImplementationAvailable", func() {
		s.Error(err)
		var noImplErr *podman.ErrNoImplementationAvailable
		s.ErrorAs(err, &noImplErr)
	})

	s.Run("error lists all tried implementations", func() {
		s.Contains(err.Error(), "mock1 (not available)")
		s.Contains(err.Error(), "mock2 (not available)")
	})
}

func (s *InterfaceSuite) TestNewPodmanAutoDetectionSelectsHighestPriority() {
	// Clear and register mocks with different priorities
	podman.Clear()

	lowPriorityCalled := false
	highPriorityCalled := false

	podman.Register(&mockImplementation{
		name:      "low-priority",
		available: true,
		priority:  10,
		initFunc: func(_ config.Config) (podman.Podman, error) {
			lowPriorityCalled = true
			return nil, nil
		},
	})
	podman.Register(&mockImplementation{
		name:      "high-priority",
		available: true,
		priority:  100,
		initFunc: func(_ config.Config) (podman.Podman, error) {
			highPriorityCalled = true
			return nil, nil
		},
	})

	cfg := config.Config{} // Auto-detect
	_, _ = podman.NewPodman(cfg)

	s.Run("calls high priority implementation", func() {
		s.True(highPriorityCalled)
	})

	s.Run("does not call low priority implementation", func() {
		s.False(lowPriorityCalled)
	})
}

func (s *InterfaceSuite) TestNewPodmanAutoDetectionSkipsUnavailable() {
	// Clear and register mocks where high priority is unavailable
	podman.Clear()

	fallbackCalled := false

	podman.Register(&mockImplementation{
		name:      "unavailable-high",
		available: false,
		priority:  100,
	})
	podman.Register(&mockImplementation{
		name:      "available-low",
		available: true,
		priority:  10,
		initFunc: func(_ config.Config) (podman.Podman, error) {
			fallbackCalled = true
			return nil, nil
		},
	})

	cfg := config.Config{} // Auto-detect
	_, err := podman.NewPodman(cfg)

	s.Run("succeeds", func() {
		s.NoError(err)
	})

	s.Run("falls back to available implementation", func() {
		s.True(fallbackCalled)
	})
}

func (s *InterfaceSuite) TestImplementationInterface() {
	impl := podman.ImplementationFromString("cli")
	if impl == nil {
		s.T().Skip("CLI implementation not registered")
	}

	s.Run("Name returns non-empty string", func() {
		s.NotEmpty(impl.Name())
	})

	s.Run("Description returns non-empty string", func() {
		s.NotEmpty(impl.Description())
	})

	s.Run("Priority returns positive value", func() {
		s.Greater(impl.Priority(), 0)
	})

	s.Run("Available returns boolean without panic", func() {
		_ = impl.Available()
	})

	s.Run("Initialize returns Podman instance when available", func() {
		if !impl.Available() {
			s.T().Skip("CLI implementation not available")
		}
		cfg := config.Default()
		p, err := impl.Initialize(cfg)
		s.NoError(err)
		s.NotNil(p)
	})
}
