package test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Must panics if err is not nil, otherwise returns v.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// ReadFile reads a file relative to the caller's source file location.
func ReadFile(path ...string) string {
	_, file, _, _ := runtime.Caller(1)
	filePath := filepath.Join(append([]string{filepath.Dir(file)}, path...)...)
	fileBytes := Must(os.ReadFile(filePath))
	return string(fileBytes)
}

// CreateTempFile creates a temporary file with the given name and content.
// The file is automatically cleaned up when the test completes.
// Returns the absolute path to the created file.
func CreateTempFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file %s: %v", name, err)
	}
	return filePath
}
