package config

import "github.com/dariubs/tower/libtower"

// TCP config type
type TCP struct {
	IP      string           `yaml:"ip"`      // IPv4 or IPv6
	Port    int              `yaml:"port"`    // Port number
	Timeout libtower.Timeout `yaml:"timeout"` // Timeout
}
