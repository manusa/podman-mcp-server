package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type RootCmdSuite struct {
	suite.Suite
}

func TestRootCmdSuite(t *testing.T) {
	suite.Run(t, new(RootCmdSuite))
}

func (s *RootCmdSuite) SetupTest() {
	viper.Reset()
	rootCmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
	_ = viper.BindPFlags(rootCmd.Flags())
}

func (s *RootCmdSuite) captureOutput(args []string) (string, error) {
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

func (s *RootCmdSuite) TestVersion() {
	version, err := s.captureOutput([]string{"--version"})

	s.NoError(err)
	s.Equal("0.0.0\n", version)
}

func (s *RootCmdSuite) TestHelpPortFlag() {
	help, err := s.captureOutput([]string{"--help"})
	s.Require().NoError(err)

	s.Run("contains port flag with short form", func() {
		s.Regexp(`-p, --port int`, help)
	})

	s.Run("contains port flag description", func() {
		s.Regexp(`--port int\s+Start HTTP server on the specified port`, help)
	})

	s.Run("describes Streamable HTTP endpoints", func() {
		s.Contains(help, "Streamable HTTP at /mcp and SSE at /sse")
	})
}

func (s *RootCmdSuite) TestHelpHidesDeprecatedFlags() {
	help, err := s.captureOutput([]string{"--help"})
	s.Require().NoError(err)

	s.Run("hides deprecated sse-port flag", func() {
		s.NotContains(help, "--sse-port")
	})

	s.Run("hides deprecated sse-base-url flag", func() {
		s.NotContains(help, "--sse-base-url")
	})
}

func (s *RootCmdSuite) TestHelpPodmanImplFlag() {
	help, err := s.captureOutput([]string{"--help"})
	s.Require().NoError(err)

	s.Run("contains podman-impl flag", func() {
		s.Regexp(`--podman-impl string`, help)
	})

	s.Run("contains flag description", func() {
		s.Regexp(`--podman-impl string\s+Podman implementation to use`, help)
	})

	s.Run("lists available implementations", func() {
		s.Regexp(`\(available: [a-z, ]+\)`, help)
	})

	s.Run("includes cli in available implementations", func() {
		s.Regexp(`\(available:.*cli.*\)`, help)
	})

	s.Run("mentions auto-detection", func() {
		s.Regexp(`Auto-detects if not specified`, help)
	})
}
