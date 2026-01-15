package test

import (
	"os"
	"path/filepath"
	"runtime"
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
