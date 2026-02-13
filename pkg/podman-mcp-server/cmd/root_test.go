package cmd

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func() error) (string, error) {
	originalOut := os.Stdout
	defer func() {
		os.Stdout = originalOut
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := f()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func TestVersion(t *testing.T) {
	rootCmd.SetArgs([]string{"--version"})
	version, err := captureOutput(rootCmd.Execute)
	if version != "0.0.0\n" {
		t.Fatalf("Expected version 0.0.0, got %s %v", version, err)
		return
	}
}

func TestHelpContainsPortFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	help, err := captureOutput(rootCmd.Execute)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !strings.Contains(help, "--port") {
		t.Fatalf("Expected help to contain --port flag, got %s", help)
	}
	if !strings.Contains(help, "Streamable HTTP at /mcp and SSE at /sse") {
		t.Fatalf("Expected help to contain Streamable HTTP endpoint description, got %s", help)
	}
}

func TestHelpHidesDeprecatedFlags(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	help, err := captureOutput(rootCmd.Execute)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Deprecated flags should be hidden from help output
	if strings.Contains(help, "--sse-port") {
		t.Fatalf("Expected help to NOT contain deprecated --sse-port flag, got %s", help)
	}
	if strings.Contains(help, "--sse-base-url") {
		t.Fatalf("Expected help to NOT contain deprecated --sse-base-url flag, got %s", help)
	}
}

func TestHelpContainsPodmanImplFlag(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	help, err := captureOutput(rootCmd.Execute)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !strings.Contains(help, "--podman-impl") {
		t.Fatalf("Expected help to contain --podman-impl flag, got %s", help)
	}
	// Check that available implementations are listed
	if !strings.Contains(help, "available:") {
		t.Fatalf("Expected help to list available implementations, got %s", help)
	}
	if !strings.Contains(help, "cli") {
		t.Fatalf("Expected help to list 'cli' implementation, got %s", help)
	}
}
