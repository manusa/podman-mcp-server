package podman_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/pkg/podman"
)

type PodmanCliSuite struct {
	suite.Suite
	savedImplementations []podman.Implementation
	originalPath         string
}

func TestPodmanCli(t *testing.T) {
	suite.Run(t, new(PodmanCliSuite))
}

func (s *PodmanCliSuite) SetupTest() {
	// Save current registry state
	s.savedImplementations = podman.Implementations()
	// Save original PATH
	s.originalPath = os.Getenv("PATH")
}

func (s *PodmanCliSuite) TearDownTest() {
	// Restore PATH
	_ = os.Setenv("PATH", s.originalPath)
	// Restore registry state
	podman.Clear()
	for _, impl := range s.savedImplementations {
		podman.Register(impl)
	}
}

// createMockBinariesInPath creates mock executables and sets PATH to their directory.
// On Unix, it creates shell scripts that exit successfully.
// On Windows, it creates batch files (.cmd) that exit successfully.
// Returns the temp directory path (auto-cleaned up by t.TempDir()).
func (s *PodmanCliSuite) createMockBinariesInPath(names ...string) string {
	tmpDir := s.T().TempDir()
	for _, name := range names {
		var content string
		var path string

		if runtime.GOOS == "windows" {
			// Strip any extension and use .cmd for batch files
			baseName := name
			if ext := filepath.Ext(name); ext != "" {
				baseName = name[:len(name)-len(ext)]
			}
			path = filepath.Join(tmpDir, baseName+".cmd")
			// Batch file that exits successfully
			content = "@exit /b 0\r\n"
		} else {
			path = filepath.Join(tmpDir, name)
			// Shell script that exits successfully
			content = "#!/bin/sh\nexit 0\n"
		}

		err := os.WriteFile(path, []byte(content), 0755)
		s.Require().NoError(err)
	}
	_ = os.Setenv("PATH", tmpDir)
	return tmpDir
}

func (s *PodmanCliSuite) TestCLIImplementationMetadata() {
	impl := podman.ImplementationFromString("cli")
	s.Require().NotNil(impl, "CLI implementation must be registered")

	s.Run("Name returns cli", func() {
		s.Equal("cli", impl.Name())
	})

	s.Run("Description returns Podman CLI wrapper", func() {
		s.Equal("Podman CLI wrapper", impl.Description())
	})

	s.Run("Priority returns 50", func() {
		s.Equal(50, impl.Priority())
	})
}

func (s *PodmanCliSuite) TestCLIAvailable() {
	impl := podman.ImplementationFromString("cli")
	s.Require().NotNil(impl)

	s.Run("returns true when podman is in PATH", func() {
		s.createMockBinariesInPath("podman")
		s.True(impl.Available())
	})

	s.Run("returns true when podman.exe is in PATH", func() {
		s.createMockBinariesInPath("podman.exe")
		s.True(impl.Available())
	})

	s.Run("returns false when no podman binary in PATH", func() {
		s.createMockBinariesInPath() // Empty PATH
		s.False(impl.Available())
	})
}

func (s *PodmanCliSuite) TestCLINew() {
	impl := podman.ImplementationFromString("cli")
	s.Require().NotNil(impl)

	s.Run("succeeds when podman is in PATH", func() {
		s.createMockBinariesInPath("podman")
		p, err := impl.New()
		s.NoError(err)
		s.NotNil(p)
	})

	s.Run("succeeds when podman.exe is in PATH", func() {
		s.createMockBinariesInPath("podman.exe")
		p, err := impl.New()
		s.NoError(err)
		s.NotNil(p)
	})

	s.Run("succeeds when both podman and podman.exe exist", func() {
		s.createMockBinariesInPath("podman", "podman.exe")
		p, err := impl.New()
		s.NoError(err)
		s.NotNil(p)
	})

	s.Run("fails when no binary is in PATH", func() {
		s.createMockBinariesInPath() // Empty PATH
		_, err := impl.New()
		s.Error(err)
		s.Contains(err.Error(), "podman CLI not found")
	})

	s.Run("fails when binary exists but version command fails", func() {
		tmpDir := s.createMockBinariesInPath() // Sets PATH to tmpDir
		// Create a binary that fails (exits with non-zero)
		var binaryPath, content string
		if runtime.GOOS == "windows" {
			binaryPath = filepath.Join(tmpDir, "podman.cmd")
			content = "@exit /b 1\r\n"
		} else {
			binaryPath = filepath.Join(tmpDir, "podman")
			content = "#!/bin/sh\nexit 1\n"
		}
		err := os.WriteFile(binaryPath, []byte(content), 0755)
		s.Require().NoError(err)

		_, err = impl.New()
		s.Error(err)
		s.Contains(err.Error(), "podman CLI not found")
	})
}

func (s *PodmanCliSuite) TestNewPodmanWithCLI() {
	s.Run("empty override returns CLI implementation", func() {
		s.createMockBinariesInPath("podman")
		p, err := podman.NewPodman("")
		s.NoError(err)
		s.NotNil(p)
	})

	s.Run("explicit cli override returns CLI implementation", func() {
		s.createMockBinariesInPath("podman")
		p, err := podman.NewPodman("cli")
		s.NoError(err)
		s.NotNil(p)
	})

	s.Run("cli override fails when podman not available", func() {
		s.createMockBinariesInPath() // Empty PATH
		_, err := podman.NewPodman("cli")
		s.Error(err)

		var notAvailErr *podman.ErrImplementationNotAvailable
		s.ErrorAs(err, &notAvailErr)
		s.Equal("cli", notAvailErr.Name)
	})
}

func (s *PodmanCliSuite) TestNewPodmanAutoDetectionWithCLI() {
	s.Run("auto-detects CLI when available", func() {
		s.createMockBinariesInPath("podman")
		p, err := podman.NewPodman()
		s.NoError(err)
		s.NotNil(p)
	})

	s.Run("returns error when CLI not available and no other implementations", func() {
		s.createMockBinariesInPath() // Empty PATH
		_, err := podman.NewPodman()
		s.Error(err)

		var noImplErr *podman.ErrNoImplementationAvailable
		s.ErrorAs(err, &noImplErr)
		s.Contains(err.Error(), "cli (not available)")
	})
}

func (s *PodmanCliSuite) TestCLINewWithRealBinary() {
	if runtime.GOOS != "linux" {
		s.T().Skip("real podman binary tests only run on Linux")
	}

	// This test verifies behavior with the actual podman binary if available
	impl := podman.ImplementationFromString("cli")
	s.Require().NotNil(impl)

	// Restore original PATH for this test
	_ = os.Setenv("PATH", s.originalPath)

	s.Run("works with real podman binary", func() {
		// Check if real podman is available
		_, err := exec.LookPath("podman")
		if err != nil {
			s.T().Skip("real podman binary not available")
		}

		p, err := impl.New()
		s.NoError(err)
		s.NotNil(p)
	})
}
