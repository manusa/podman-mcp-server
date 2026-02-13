package test

import (
	"net/url"
	"os"
	"os/exec"
	"testing"
)

// IsPodmanAvailable checks if the real podman binary is available and functional.
// This uses the same check as newPodmanCli() to ensure consistency - if this returns
// true, NewPodman() should succeed.
func IsPodmanAvailable() bool {
	for _, cmd := range []string{"podman", "podman.exe"} {
		filePath, err := exec.LookPath(cmd)
		if err != nil {
			continue
		}
		// Use "version" subcommand (not --version flag) to match newPodmanCli() behavior.
		// On macOS/Windows, this requires a running podman machine, which is intentional -
		// we want to skip tests when podman isn't fully functional.
		if _, err = exec.Command(filePath, "version").CombinedOutput(); err == nil {
			return true
		}
	}
	return false
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
