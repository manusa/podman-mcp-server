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
// Note: Environment is restored by RestoreEnv in TearDownTest.
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
