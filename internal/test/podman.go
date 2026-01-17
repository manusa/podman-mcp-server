package test

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"
)

// WithPodmanBinary builds a fake podman binary and prepends its directory to PATH.
// Returns the directory containing the fake binary.
// The fake binary echoes its arguments and can optionally return content from output.txt.
func WithPodmanBinary(t *testing.T) string {
	binDir := t.TempDir()
	binary := "podman"
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	// Build the fake podman binary from testdata
	output, err := exec.
		Command("go", "build", "-o", path.Join(binDir, binary),
			path.Join(testdataPath(), "podman", "main.go")).
		CombinedOutput()
	if err != nil {
		panic(fmt.Errorf("failed to generate podman binary: %w, output: %s", err, string(output)))
	}
	if os.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH")) != nil {
		panic("failed to set PATH")
	}
	return binDir
}

// testdataPath returns the path to the testdata directory.
func testdataPath() string {
	// Navigate from internal/test to project root, then to testdata
	return path.Join("..", "..", "testdata")
}

// IsPodmanAvailable checks if the real podman binary is available in PATH.
func IsPodmanAvailable() bool {
	filePath, err := exec.LookPath("podman")
	if err != nil {
		return false
	}
	// Verify it's actually podman and not our fake binary
	output, err := exec.Command(filePath, "version", "--format", "{{.Client.Version}}").CombinedOutput()
	if err != nil {
		return false
	}
	// Real podman outputs a version number like "4.0.0"
	version := strings.TrimSpace(string(output))
	return len(version) > 0 && version[0] >= '0' && version[0] <= '9'
}

// IsDockerAvailable checks if the docker binary is available in PATH.
func IsDockerAvailable() bool {
	filePath, err := exec.LookPath("docker")
	if err != nil {
		return false
	}
	output, err := exec.Command(filePath, "version", "--format", "{{.Client.Version}}").CombinedOutput()
	if err != nil {
		return false
	}
	version := strings.TrimSpace(string(output))
	return len(version) > 0 && version[0] >= '0' && version[0] <= '9'
}

// WithContainerHost sets the CONTAINER_HOST environment variable to point to the mock server.
// This redirects podman CLI commands to the mock server.
// Returns a cleanup function that restores the original environment.
func WithContainerHost(t *testing.T, serverURL string) func() {
	originalContainerHost := os.Getenv("CONTAINER_HOST")
	originalDockerHost := os.Getenv("DOCKER_HOST")

	// Convert http://localhost:PORT to tcp://localhost:PORT for podman
	u, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}
	tcpURL := "tcp://" + u.Host

	if err := os.Setenv("CONTAINER_HOST", tcpURL); err != nil {
		t.Fatalf("failed to set CONTAINER_HOST: %v", err)
	}
	if err := os.Setenv("DOCKER_HOST", tcpURL); err != nil {
		t.Fatalf("failed to set DOCKER_HOST: %v", err)
	}

	return func() {
		if originalContainerHost != "" {
			_ = os.Setenv("CONTAINER_HOST", originalContainerHost)
		} else {
			_ = os.Unsetenv("CONTAINER_HOST")
		}
		if originalDockerHost != "" {
			_ = os.Setenv("DOCKER_HOST", originalDockerHost)
		} else {
			_ = os.Unsetenv("DOCKER_HOST")
		}
	}
}
