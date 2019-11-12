// Package config is a parser for tower config files
package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config type
type Config struct {
	TCP []TCP `yaml:"tcp"`
}

// Parse config file
func Parse(filepath string) (Config, error) {
	var c Config

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
