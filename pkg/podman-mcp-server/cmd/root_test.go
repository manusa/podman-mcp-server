package cmd

import (
	"io"
	"os"
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
