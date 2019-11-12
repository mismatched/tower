// Package config is a parser for tower config files
package config

// Config type
type Config struct {
	TCP []TCP `yaml:"tcp"`
}
