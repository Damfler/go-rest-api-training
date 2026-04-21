package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Server.Port != 999 {
		t.Errorf("cfg.Server.Port should be 999, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "localhost" {
		t.Errorf("cfg.Server.Host should be localhost, got %s", cfg.Server.Host)
	}
}

func TestLoadYaml(t *testing.T) {
	yaml := `
server:
  port: 8081
database:
  path: testing.db
`

	path := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Server.Port != 8081 {
		t.Errorf("cfg.Server.Port should be 8081, got %d", cfg.Server.Port)
	}
	if cfg.Database.Path != "testing.db" {
		t.Errorf("cfg.Database.Path should be testing.db, got %s", cfg.Database.Path)
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("cfg.Server.Host should be localhost, got %s", cfg.Server.Host)
	}
}

func TestEnvOverride(t *testing.T) {
	yaml := `
server:
  port: 8081
`

	path := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SERVER_PORT", "8080")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("cfg.Server.Port should be 8080, got %d", cfg.Server.Port)
	}
}

func TestValidation(t *testing.T) {
	yaml := `
server:
  port: 99999
`

	path := t.TempDir() + "/config.yaml"
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Errorf("expected validation error for port 99999")
	}
}
