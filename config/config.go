package config

import (
	"encoding/json"
	"encoding/xml"
	"errors"
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

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	CORS     CORSConfig     `yaml:"cors"`
	Debug    bool           `yaml:"debug"`
}

func Load(path string) (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 999,
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

	config.loadEnv()

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	var errs []error

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be 1-65535, got %d", c.Server.Port))
	}

	if c.Server.Host == "" {
		errs = append(errs, fmt.Errorf("server.host is required"))
	}

	if c.Database.Path == "" {
		errs = append(errs, fmt.Errorf("database.path is required"))
	}

	return errors.Join(errs...)
}
