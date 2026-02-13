package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutput(args []string) (string, error) {
	rootCmd.SetArgs(args)

	originalOut := os.Stdout
	defer func() {
		os.Stdout = originalOut
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := rootCmd.Execute()
	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func TestVersion(t *testing.T) {
	version, err := captureOutput([]string{"--version"})

	assert.NoError(t, err)
	assert.Equal(t, "0.0.0\n", version)
}

func TestHelpPortFlag(t *testing.T) {
	help, err := captureOutput([]string{"--help"})
	require.NoError(t, err)

	t.Run("contains port flag with short form", func(t *testing.T) {
		assert.Regexp(t, `-p, --port int`, help)
	})

	t.Run("contains port flag description", func(t *testing.T) {
		assert.Regexp(t, `--port int\s+Start HTTP server on the specified port`, help)
	})

	t.Run("describes Streamable HTTP endpoints", func(t *testing.T) {
		assert.Contains(t, help, "Streamable HTTP at /mcp and SSE at /sse")
	})
}

func TestHelpHidesDeprecatedFlags(t *testing.T) {
	help, err := captureOutput([]string{"--help"})
	require.NoError(t, err)

	t.Run("hides deprecated sse-port flag", func(t *testing.T) {
		assert.NotContains(t, help, "--sse-port")
	})

	t.Run("hides deprecated sse-base-url flag", func(t *testing.T) {
		assert.NotContains(t, help, "--sse-base-url")
	})
}

func TestHelpPodmanImplFlag(t *testing.T) {
	help, err := captureOutput([]string{"--help"})
	require.NoError(t, err)

	t.Run("contains podman-impl flag", func(t *testing.T) {
		assert.Regexp(t, `--podman-impl string`, help)
	})

	t.Run("contains flag description", func(t *testing.T) {
		assert.Regexp(t, `--podman-impl string\s+Podman implementation to use`, help)
	})

	t.Run("lists available implementations", func(t *testing.T) {
		assert.Regexp(t, `\(available: [a-z, ]+\)`, help)
	})

	t.Run("includes cli in available implementations", func(t *testing.T) {
		assert.Regexp(t, `\(available:.*cli.*\)`, help)
	})

	t.Run("mentions auto-detection", func(t *testing.T) {
		assert.Regexp(t, `Auto-detects if not specified`, help)
	})
}
