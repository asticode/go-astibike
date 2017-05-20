package astidarksky

import "flag"

// Flags
var (
	APIKey = flag.String("dark-sky-api-key", "", "the Dark Sky API key")
)

// Configuration represents the ffmpeg configuration
type Configuration struct {
	APIKey string `toml:"api_key"`
}

// FlagConfig generates a Configuration based on flags
func FlagConfig() Configuration {
	return Configuration{
		APIKey: *APIKey,
	}
}
