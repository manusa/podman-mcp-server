package test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
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
