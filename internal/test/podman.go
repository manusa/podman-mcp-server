package test

import (
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// IsPodmanAvailable checks if the real podman binary is available in PATH.
func IsPodmanAvailable() bool {
	filePath, err := exec.LookPath("podman")
	if err != nil {
		return false
	}
	// Verify it's actually podman and not our fake binary
	// Use --version instead of version subcommand to avoid needing a running machine/daemon
	output, err := exec.Command(filePath, "--version").CombinedOutput()
	if err != nil {
		return false
	}
	// Real podman outputs "podman version X.Y.Z" (may include additional info after)
	// Our fake binary would output something different
	outputStr := strings.TrimSpace(string(output))
	return strings.HasPrefix(outputStr, "podman version ")
}

// WithContainerHost sets the CONTAINER_HOST environment variable to point to the mock server.
// This redirects podman CLI commands to the mock server.
// Note: Environment is restored by RestoreEnv in TearDownTest.
func WithContainerHost(t *testing.T, serverURL string) {
	// Convert http://localhost:PORT to tcp://localhost:PORT for podman
	u, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}
	tcpURL := "tcp://" + u.Host

	if err := os.Setenv("CONTAINER_HOST", tcpURL); err != nil {
		t.Fatalf("failed to set CONTAINER_HOST: %v", err)
	}
}
