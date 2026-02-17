package config_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/pkg/config"
)

type ConfigSuite struct {
	suite.Suite
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}

func (s *ConfigSuite) TestDefault() {
	cfg := config.Default()

	s.Run("PodmanImpl is empty for auto-detect", func() {
		s.Empty(cfg.PodmanImpl)
	})

	s.Run("OutputFormat is text", func() {
		s.Equal("text", cfg.OutputFormat)
	})
}

func (s *ConfigSuite) TestWithOverrides() {
	s.Run("empty overrides returns defaults", func() {
		cfg := config.WithOverrides(config.Config{})
		s.Empty(cfg.PodmanImpl)
		s.Equal("text", cfg.OutputFormat)
	})

	s.Run("PodmanImpl override is applied", func() {
		cfg := config.WithOverrides(config.Config{PodmanImpl: "cli"})
		s.Equal("cli", cfg.PodmanImpl)
	})

	s.Run("OutputFormat text override is applied", func() {
		cfg := config.WithOverrides(config.Config{OutputFormat: "text"})
		s.Equal("text", cfg.OutputFormat)
	})

	s.Run("OutputFormat json override is applied", func() {
		cfg := config.WithOverrides(config.Config{OutputFormat: "json"})
		s.Equal("json", cfg.OutputFormat)
	})

	s.Run("invalid OutputFormat is ignored", func() {
		cfg := config.WithOverrides(config.Config{OutputFormat: "xml"})
		s.Equal("text", cfg.OutputFormat)
	})

	s.Run("multiple overrides are applied", func() {
		cfg := config.WithOverrides(config.Config{
			PodmanImpl:   "api",
			OutputFormat: "json",
		})
		s.Equal("api", cfg.PodmanImpl)
		s.Equal("json", cfg.OutputFormat)
	})

	s.Run("partial overrides preserve defaults", func() {
		cfg := config.WithOverrides(config.Config{PodmanImpl: "cli"})
		s.Equal("cli", cfg.PodmanImpl)
		s.Equal("text", cfg.OutputFormat)
	})
}

func (s *ConfigSuite) TestOutputFormatConstants() {
	s.Run("OutputFormatText value", func() {
		s.Equal("text", config.OutputFormatText)
	})

	s.Run("OutputFormatJSON value", func() {
		s.Equal("json", config.OutputFormatJSON)
	})
}
