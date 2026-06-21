// Package config is a parser for tower config files
package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// CheckConfig is a single check entry in a tower config file.
// The Type field determines which libtower check to run.
type CheckConfig struct {
	Type string `yaml:"type"`

	// TCP / TLS
	IP   string `yaml:"ip"`
	Port int    `yaml:"port"`

	// HTTPS
	Host                 string        `yaml:"host"`
	WarnIfExpiringWithin time.Duration `yaml:"warn_if_expiring"`
	InsecureSkipVerify   bool          `yaml:"insecure_skip_verify"`

	// DNS
	Addr string `yaml:"addr"`

	// Ping
	Count int `yaml:"count"`

	// HTTP / Trace
	URL    string `yaml:"url"`
	Method string `yaml:"method"`

	// Common
	Timeout time.Duration `yaml:"timeout"`
}

// Config is the root tower config file.
type Config struct {
	Checks []CheckConfig `yaml:"checks"`
}

// Parse config file
func Parse(filepath string) (Config, error) {
	var c Config

	data, err := os.ReadFile(filepath)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
