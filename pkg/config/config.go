package config

// Output format constants for list commands.
const (
	// OutputFormatText uses podman's default table format (human-readable).
	OutputFormatText = "text"
	// OutputFormatJSON outputs JSON for programmatic parsing.
	OutputFormatJSON = "json"
)

// Config holds configuration for the podman-mcp-server.
type Config struct {
	// PodmanImpl specifies which Podman implementation to use.
	// Valid values: "cli", "api". Empty string means auto-detect.
	PodmanImpl string

	// OutputFormat specifies the output format for list commands.
	// Valid values: OutputFormatText, OutputFormatJSON.
	OutputFormat string
}
