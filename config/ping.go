package config

import "github.com/mismatched/libtower"

// PING config type
type PING struct {
	Host    string           `yaml:"host"`    // IPv4 or IPv6
	Count   int              `yaml:"count"`   // Number of pings
	Timeout libtower.Timeout `yaml:"timeout"` // Timeout
}
