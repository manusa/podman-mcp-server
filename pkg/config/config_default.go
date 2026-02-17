package config

// Default returns a Config with default values.
func Default() Config {
	return Config{
		PodmanImpl:   "",               // Auto-detect
		OutputFormat: OutputFormatText, // Human-readable table format
	}
}

// WithOverrides returns a Config with default values, overridden by any
// non-empty values in the provided overrides.
// Invalid OutputFormat values are ignored (default is used).
func WithOverrides(overrides Config) Config {
	cfg := Default()
	if overrides.PodmanImpl != "" {
		cfg.PodmanImpl = overrides.PodmanImpl
	}
	if overrides.OutputFormat == OutputFormatText || overrides.OutputFormat == OutputFormatJSON {
		cfg.OutputFormat = overrides.OutputFormat
	}
	return cfg
}
