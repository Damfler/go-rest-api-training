package config

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Debug    bool           `yaml:"debug"`
}

func Load(path string) (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Path: "./app.db",
		},
		Debug: false,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}

	switch filepath.Ext(path) {
	case ".json":
		err = json.Unmarshal(data, config)
	case ".yml", ".yaml":
		err = yaml.Unmarshal(data, config)
	case ".xml":
		err = xml.Unmarshal(data, config)
	default:
		return nil, fmt.Errorf("Unsupported file type: %s\n", filepath.Ext(path))
	}

	if err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", path, err)
	}

	return config, nil
}
