package podman_test

import (
	"os"
	"path/filepath"
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

func (s *PodmanCliSuite) TestNewPodmanCliNotFound() {
	originalEnv := os.Environ()
	defer test.RestoreEnv(originalEnv)
	s.Require().NoError(os.Setenv("PATH", filepath.Join(os.TempDir(), "nonexistent-path-for-testing")))

	_, err := podman.NewPodman()

	s.Error(err, "should return an error when podman CLI is not found")
	s.Contains(err.Error(), "podman CLI not found")
}
