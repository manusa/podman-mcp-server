package podman_test

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/pkg/podman"
)

type SocketSuite struct {
	suite.Suite
	originalEnv map[string]string
}

func TestSocket(t *testing.T) {
	suite.Run(t, new(SocketSuite))
}

func (s *SocketSuite) SetupTest() {
	// Save environment variables
	s.originalEnv = map[string]string{
		"CONTAINER_HOST":  os.Getenv("CONTAINER_HOST"),
		"XDG_RUNTIME_DIR": os.Getenv("XDG_RUNTIME_DIR"),
	}
}

func (s *SocketSuite) TearDownTest() {
	// Restore environment variables
	for k, v := range s.originalEnv {
		if v == "" {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, v)
		}
	}
}

func (s *SocketSuite) TestDetectSocket() {
	s.Run("returns CONTAINER_HOST when set", func() {
		expected := "unix:///custom/path/podman.sock"
		s.Require().NoError(os.Setenv("CONTAINER_HOST", expected))
		defer func() { _ = os.Unsetenv("CONTAINER_HOST") }()

		result, err := podman.DetectSocket()

		s.NoError(err)
		s.Equal(expected, result)
	})

	s.Run("returns CONTAINER_HOST for tcp scheme", func() {
		expected := "tcp://localhost:8080"
		s.Require().NoError(os.Setenv("CONTAINER_HOST", expected))
		defer func() { _ = os.Unsetenv("CONTAINER_HOST") }()

		result, err := podman.DetectSocket()

		s.NoError(err)
		s.Equal(expected, result)
	})

	s.Run("checks XDG_RUNTIME_DIR socket path", func() {
		// Unset CONTAINER_HOST to test file detection
		_ = os.Unsetenv("CONTAINER_HOST")

		// Create a temporary directory with a socket file
		tmpDir := s.T().TempDir()
		socketDir := filepath.Join(tmpDir, "podman")
		s.Require().NoError(os.MkdirAll(socketDir, 0755))
		socketPath := filepath.Join(socketDir, "podman.sock")

		// Create a dummy file at the socket path
		f, err := os.Create(socketPath)
		s.Require().NoError(err)
		_ = f.Close()

		// Set XDG_RUNTIME_DIR
		s.Require().NoError(os.Setenv("XDG_RUNTIME_DIR", tmpDir))

		result, err := podman.DetectSocket()

		s.NoError(err)
		s.Equal("unix://"+socketPath, result)
	})

	s.Run("returns error when no socket found", func() {
		// Clear environment
		_ = os.Unsetenv("CONTAINER_HOST")
		_ = os.Setenv("XDG_RUNTIME_DIR", "/nonexistent/path")

		// This test only works if no rootful socket exists
		if _, err := os.Stat("/run/podman/podman.sock"); err == nil {
			s.T().Skip("rootful podman socket exists, skipping test")
		}
		// Check rootless fallback
		uid := os.Getuid()
		fallbackPath := "/run/user/" + strconv.Itoa(uid) + "/podman/podman.sock"
		if _, err := os.Stat(fallbackPath); err == nil {
			s.T().Skip("rootless podman socket exists, skipping test")
		}

		_, err := podman.DetectSocket()

		s.ErrorIs(err, podman.ErrNoSocketFound)
	})
}

func (s *SocketSuite) TestPingSocket() {
	s.Run("returns error for invalid URI", func() {
		err := podman.PingSocket("://invalid")

		s.Error(err)
		s.Contains(err.Error(), "invalid socket URI")
	})

	s.Run("returns error for unsupported scheme", func() {
		err := podman.PingSocket("http://localhost:8080")

		s.Error(err)
		s.Contains(err.Error(), "unsupported socket scheme")
	})

	s.Run("returns error for non-existent socket", func() {
		err := podman.PingSocket("unix:///nonexistent/podman.sock")

		s.Error(err)
		s.Contains(err.Error(), "failed to connect")
	})

	s.Run("succeeds with valid unix socket", func() {
		if runtime.GOOS != "linux" {
			s.T().Skip("Unix socket tests only run on Linux")
		}
		// Create a temporary socket
		tmpDir := s.T().TempDir()
		socketPath := filepath.Join(tmpDir, "test.sock")

		// Create a mock server
		listener, err := net.Listen("unix", socketPath)
		s.Require().NoError(err)
		defer func() { _ = listener.Close() }()

		// Handle requests
		server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})}
		go func() { _ = server.Serve(listener) }()
		defer func() { _ = server.Close() }()

		err = podman.PingSocket("unix://" + socketPath)

		s.NoError(err)
	})

	s.Run("succeeds with valid tcp socket", func() {
		// Create a mock TCP server
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		s.Require().NoError(err)
		defer func() { _ = listener.Close() }()

		// Handle requests
		server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})}
		go func() { _ = server.Serve(listener) }()
		defer func() { _ = server.Close() }()

		err = podman.PingSocket("tcp://" + listener.Addr().String())

		s.NoError(err)
	})

	s.Run("returns error for non-Podman endpoint", func() {
		if runtime.GOOS != "linux" {
			s.T().Skip("Unix socket tests only run on Linux")
		}
		// Create a temporary socket that returns a non-200 response
		tmpDir := s.T().TempDir()
		socketPath := filepath.Join(tmpDir, "test.sock")

		listener, err := net.Listen("unix", socketPath)
		s.Require().NoError(err)
		defer func() { _ = listener.Close() }()

		// Handle requests with 404
		server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.NotFound(w, nil)
		})}
		go func() { _ = server.Serve(listener) }()
		defer func() { _ = server.Close() }()

		err = podman.PingSocket("unix://" + socketPath)

		s.Error(err)
		s.Contains(err.Error(), "unexpected ping response")
	})
}
