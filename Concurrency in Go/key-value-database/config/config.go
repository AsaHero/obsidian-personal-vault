package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Engine struct {
		Type string `yaml:"type"`
	} `yaml:"engine"`
	Network struct {
		Address        string `yaml:"address"`
		MaxConnections int    `yaml:"max_connections"`
		MaxMessageSize string `yaml:"max_message_size"`
		IdleTimeout    string `yaml:"idle_timeout"`
	} `yaml:"network"`
	Logging struct {
		Level  string `yaml:"level"`
		Output string `yaml:"output"`
	} `yaml:"logging"`
}

func DefaultConfig() *Config {
	return &Config{
		Engine: struct {
			Type string "yaml:\"type\""
		}{
			Type: "in_memory",
		},
		Network: struct {
			Address        string "yaml:\"address\""
			MaxConnections int    "yaml:\"max_connections\""
			MaxMessageSize string "yaml:\"max_message_size\""
			IdleTimeout    string "yaml:\"idle_timeout\""
		}{
			Address:        "127.0.0.1:3223",
			MaxConnections: 100,
			MaxMessageSize: "4KB",
			IdleTimeout:    "5m",
		},
		Logging: struct {
			Level  string "yaml:\"level\""
			Output string "yaml:\"output\""
		}{
			Level:  "info",
			Output: "/log/output.log",
		},
	}
}

func New(path string) (*Config, error) {
	cfg := DefaultConfig()

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config yaml: %w", err)
	}

	if err := yaml.NewDecoder(file).Decode(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config yaml: %w", err)
	}

	return cfg, nil
}
