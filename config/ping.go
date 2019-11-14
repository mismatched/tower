package config

// PING config type
type PING struct {
	Host  string `yaml:"ip"`   // IPv4 or IPv6
	Count int    `yaml:"port"` // Number of pings
}
