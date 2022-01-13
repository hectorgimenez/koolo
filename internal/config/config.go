package config

import (
	"fmt"
	"gopkg.in/ini.v1"
)

type Config struct {
	Display int `ini:"display"`
}

// Load reads the config.ini file and returns a Config struct filled with data from the ini file
func Load() (Config, error) {
	f, err := ini.Load("../../config/config.ini")
	if err != nil {
		return Config{}, fmt.Errorf("error loading config.ini: %w", err)
	}

	cfg := Config{}
	if err = f.MapTo(&cfg); err != nil {
		return Config{}, fmt.Errorf("error reading config: %w", err)
	}

	return cfg, nil
}
