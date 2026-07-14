// Package config loads chater runtime configuration from the environment only
// (Go canon: env-only config, no flags, no files).
package config

import (
	"net"
	"os"
)

// Config holds runtime configuration. Every field is sourced from the
// environment via Load.
type Config struct {
	// Addr is the TCP bind address for the HTTP server, as host:port.
	Addr string
}

const (
	// envPort selects the listen port.
	envPort = "CHATER_PORT"

	// defaultPort is a local-dev fallback only. The production port (8020,
	// reserved in devopser registry/ports.md) is injected via CHATER_PORT at
	// deploy time — no deployment port is hardcoded in the binary.
	defaultPort = "8080"
)

// Load reads configuration from the environment. It returns an error type for
// forward-compatibility (validation grows in later steps); today it cannot fail.
func Load() (Config, error) {
	port := os.Getenv(envPort)
	if port == "" {
		port = defaultPort
	}

	return Config{Addr: net.JoinHostPort("", port)}, nil
}
