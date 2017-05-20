package main

import (
	"flag"

	"github.com/BurntSushi/toml"
	"github.com/asticode/go-astibike/darksky"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astiredis"
	"github.com/imdario/mergo"
)

// Flags
var (
	configPath    = flag.String("config", "", "config path")
	pathStatic    = flag.String("path-static", "", "the static path")
	pathTemplates = flag.String("path-templates", "", "the templates path")
	serverAddr    = flag.String("server-addr", "", "the server addr")
)

// Configuration represents a configuration
type Configuration struct {
	DarkSky       astidarksky.Configuration `toml:"dark_sky"`
	Logger        astilog.Configuration     `toml:"logger"`
	PathStatic    string                    `toml:"path_static"`
	PathTemplates string                    `toml:"path_templates"`
	Redis         astiredis.Configuration   `toml:"redis"`
	ServerAddr    string                    `toml:"server_addr"` // Should be of the form host:port
}

// TOMLDecodeFile allows testing functions using it
var TOMLDecodeFile = func(fpath string, v interface{}) (toml.MetaData, error) {
	return toml.DecodeFile(fpath, v)
}

// NewConfiguration creates a new configuration object
func NewConfiguration() Configuration {
	// Global config
	var gc = Configuration{
		Logger: astilog.Configuration{
			AppName: "astibike",
		},
		PathStatic:    "static",
		PathTemplates: "templates",
		Redis: astiredis.Configuration{
			Prefix: "astibike",
		},
	}

	// Local config
	if *configPath != "" {
		// Decode local config
		if _, err := TOMLDecodeFile(*configPath, &gc); err != nil {
			astilog.Fatalf("%v while decoding the config path %s", err, *configPath)
		}
	}

	// Flag config
	var c = Configuration{
		DarkSky:       astidarksky.FlagConfig(),
		Logger:        astilog.FlagConfig(),
		PathStatic:    *pathStatic,
		PathTemplates: *pathTemplates,
		Redis:         astiredis.FlagConfig(),
		ServerAddr:    *serverAddr,
	}

	// Merge configs
	if err := mergo.Merge(&c, gc); err != nil {
		astilog.Fatalf("%v while merging configs", err)
	}

	// Return
	return c
}
