package podman_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/internal/test"
	"github.com/manusa/podman-mcp-server/pkg/podman"
)

type PodmanCliSuite struct {
	suite.Suite
}

func TestPodmanCli(t *testing.T) {
	suite.Run(t, new(PodmanCliSuite))
}

func (s *PodmanCliSuite) SetupSuite() {
	if !test.IsPodmanAvailable() {
		// On Linux, podman should always be available - fail the test
		// On other platforms (macOS, Windows), skip if podman is not available
		if runtime.GOOS == "linux" {
			s.T().Fatal("podman CLI is not available - install podman to run these tests")
		}
		s.T().Skip("podman CLI not available (expected on non-Linux platforms without podman machine)")
	}
}

func (s *PodmanCliSuite) TestNewPodmanCliNotFound() {
	originalEnv := os.Environ()
	defer test.RestoreEnv(originalEnv)
	s.Require().NoError(os.Setenv("PATH", filepath.Join(os.TempDir(), "nonexistent-path-for-testing")))

	_, err := podman.NewPodman()

	s.Error(err, "should return an error when podman CLI is not found")
	s.Contains(err.Error(), "podman CLI not found")
}

func (s *PodmanCliSuite) TestNewPodmanWithOverride() {
	s.Run("empty override uses default implementation", func() {
		p, err := podman.NewPodman("")
		s.NoError(err)
		s.NotNil(p)
	})

	s.Run("explicit cli override returns CLI implementation", func() {
		p, err := podman.NewPodman("cli")
		s.NoError(err)
		s.NotNil(p)
	})

	s.Run("invalid override returns ErrUnknownImplementation", func() {
		_, err := podman.NewPodman("invalid-impl")
		s.Error(err)

		// Check it's the right error type
		var unknownErr *podman.ErrUnknownImplementation
		s.ErrorAs(err, &unknownErr)
		s.Equal("invalid-impl", unknownErr.Name)
		s.Contains(unknownErr.Available, "cli")
	})

	s.Run("error message lists valid options", func() {
		_, err := podman.NewPodman("foo")
		s.Error(err)
		s.Contains(err.Error(), "invalid podman implementation")
		s.Contains(err.Error(), "foo")
		s.Contains(err.Error(), "cli")
	})
}
